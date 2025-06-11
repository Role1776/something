package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"todoai/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user exists")
)

const userExistsError = "SQLSTATE 23505"

type Auth interface {
	GetByLoginAndPassword(ctx context.Context, authData *models.SecondAuth) (int, bool, error)
	CreateUser(ctx context.Context, authData *models.FirstAuth) (int, error)
	Verify(ctx context.Context, code string) error
	CreateVerificationCode(ctx context.Context, userID int, code string, expiresAt time.Time) error
	GetRefreshTokenData(ctx context.Context, refreshToken string) (models.RefreshTokenData, error)
	CreateOrUpdateRefreshToken(ctx context.Context, userID int, deviceID string, refreshToken string, expiresAt time.Time) error
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	GetUserByEmail(ctx context.Context, email string) (int, bool, error)
}

type authRepo struct {
	db *sql.DB
}

func newAuthRepo(db *sql.DB) *authRepo {
	return &authRepo{db: db}
}

func (r *authRepo) GetByLoginAndPassword(ctx context.Context, authData *models.SecondAuth) (id int, verified bool, err error) {
	const op = "repository.SignIn"
	const query = `
    SELECT id, verified FROM users WHERE login = $1 AND password = $2;
`
	err = r.db.QueryRowContext(ctx, query, authData.Login, authData.Password).Scan(&id, &verified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	return id, verified, nil
}

func (r *authRepo) CreateUser(ctx context.Context, authData *models.FirstAuth) (int, error) {
	const op = "repository.SignUp"
	prepare, err := r.db.PrepareContext(ctx, `
		INSERT INTO users (email, login, password, verified)
		VALUES ($1, $2, $3, $4) RETURNING id;
	`)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer prepare.Close()

	var id int
	err = prepare.QueryRowContext(ctx, authData.Email, authData.Login, authData.Password, false).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), userExistsError) {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *authRepo) CreateVerificationCode(ctx context.Context, userID int, code string, expiresAt time.Time) error {
	const op = "repository.CreateVerificationCode"
	const query = `
        INSERT INTO verification_codes (user_id, code, expires_at)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.ExecContext(ctx, query, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *authRepo) Verify(ctx context.Context, code string) error {
	const op = "repository.Verify"
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer tx.Rollback()

	var userID int
	const selectQuery = `
        SELECT user_id FROM verification_codes
        WHERE code = $1 AND expires_at > NOW()
		FOR UPDATE;
    `
	err = tx.QueryRowContext(ctx, selectQuery, code).Scan(&userID)
	if err != nil {
		tx.Rollback()
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	const updateUser = `UPDATE users SET verified = true WHERE id = $1`
	if _, err := tx.ExecContext(ctx, updateUser, userID); err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}
	const updateCode = `DELETE FROM verification_codes WHERE code = $1`
	if _, err := tx.ExecContext(ctx, updateCode, code); err != nil {
		tx.Rollback()
		return fmt.Errorf("%s: %w", op, err)
	}

	return tx.Commit()
}

func (r *authRepo) GetRefreshTokenData(ctx context.Context, refreshToken string) (models.RefreshTokenData, error) {
	const op = "repository.GetRefreshTokenData"

	const query = `
		SELECT user_id, device_id, expires_at FROM refresh_tokens WHERE refresh_token = $1;
	`
	var userID int
	var deviceID string
	var expiresAt time.Time
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(&userID, &deviceID, &expiresAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.RefreshTokenData{}, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return models.RefreshTokenData{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.RefreshTokenData{
		UserID:    userID,
		DeviceID:  deviceID,
		ExpiresAt: expiresAt,
	}, nil
}

func (r *authRepo) CreateOrUpdateRefreshToken(ctx context.Context, userID int, deviceID string, refreshToken string, expiresAt time.Time) error {
	const op = "repository.CreateOrUpdateRefreshToken"

	const query = `
        INSERT INTO refresh_tokens (user_id, device_id, refresh_token, expires_at)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id, device_id) DO UPDATE
        SET refresh_token = EXCLUDED.refresh_token,
            expires_at = EXCLUDED.expires_at,
			created_at = EXCLUDED.created_at
    `

	_, err := r.db.ExecContext(ctx, query, userID, deviceID, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *authRepo) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
	const op = "repository.DeleteRefreshToken"

	const query = `
		DELETE FROM refresh_tokens WHERE refresh_token = $1;
	`
	res, err := r.db.ExecContext(ctx, query, refreshToken)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, ErrUserNotFound)
	}

	return nil
}

func (r *authRepo) GetUserByEmail(ctx context.Context, email string) (int, bool, error) {
	const op = "repository.GetUserByEmail"
	const query = `
		SELECT id, verified FROM users WHERE email = $1;
	`
	var id int
	var verified bool
	err := r.db.QueryRowContext(ctx, query, email).Scan(&id, &verified)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return 0, false, fmt.Errorf("%s: %w", op, err)
	}

	return id, verified, nil
}
