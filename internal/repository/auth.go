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
	GetUserIDByVerificationCode(ctx context.Context, code string) (int, error)
	SetUserVerified(ctx context.Context, userID int) error
	DeleteVerificationCode(ctx context.Context, code string) error
	UpsertVerificationCode(ctx context.Context, userID int, code string, expiresAt time.Time) error
	GetRefreshTokenData(ctx context.Context, refreshToken string) (models.RefreshTokenData, error)
	CreateOrUpdateRefreshToken(ctx context.Context, userID int, deviceID string, refreshToken string, expiresAt time.Time) error
	DeleteRefreshToken(ctx context.Context, refreshToken string) error
	GetUserByEmail(ctx context.Context, email string) (int, bool, error)
}

type AuthRepo struct {
	db Querier
}

func NewAuthRepo(db Querier) *AuthRepo {
	return &AuthRepo{db: db}
}

func (r *AuthRepo) GetByLoginAndPassword(ctx context.Context, authData *models.SecondAuth) (id int, verified bool, err error) {
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

func (r *AuthRepo) CreateUser(ctx context.Context, authData *models.FirstAuth) (int, error) {
	const op = "repository.SignUp"
	const query = `
		INSERT INTO users (email, login, password, verified)
		VALUES ($1, $2, $3, $4) RETURNING id;
	`
	var id int
	err := r.db.QueryRowContext(ctx, query, authData.Email, authData.Login, authData.Password, false).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), userExistsError) {
			return 0, fmt.Errorf("%s: %w", op, ErrUserExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (r *AuthRepo) UpsertVerificationCode(ctx context.Context, userID int, code string, expiresAt time.Time) error {
	const query = `
        INSERT INTO verification_codes (user_id, code, expires_at)
        VALUES ($1, $2, $3)
        ON CONFLICT (user_id) DO UPDATE 
        SET code = EXCLUDED.code, 
            expires_at = EXCLUDED.expires_at
    `
	_, err := r.db.ExecContext(ctx, query, userID, code, expiresAt)
	if err != nil {
		return fmt.Errorf("repository.UpsertVerificationCode: %w", err)
	}
	return nil
}

func (r *AuthRepo) GetUserIDByVerificationCode(ctx context.Context, code string) (int, error) {
	const op = "repository.GetUserIDByVerificationCode"
	var userID int
	const query = `
        SELECT user_id FROM verification_codes
        WHERE code = $1 AND expires_at > NOW()
        FOR UPDATE;
    `
	err := r.db.QueryRowContext(ctx, query, code).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return userID, nil
}

func (r *AuthRepo) SetUserVerified(ctx context.Context, userID int) error {
	const op = "repository.SetUserVerified"
	const query = `UPDATE users SET verified = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *AuthRepo) DeleteVerificationCode(ctx context.Context, code string) error {
	const op = "repository.DeleteVerificationCode"
	const query = `DELETE FROM verification_codes WHERE code = $1`
	_, err := r.db.ExecContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *AuthRepo) GetRefreshTokenData(ctx context.Context, refreshToken string) (models.RefreshTokenData, error) {
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

func (r *AuthRepo) CreateOrUpdateRefreshToken(ctx context.Context, userID int, deviceID string, refreshToken string, expiresAt time.Time) error {
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

func (r *AuthRepo) DeleteRefreshToken(ctx context.Context, refreshToken string) error {
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

func (r *AuthRepo) GetUserByEmail(ctx context.Context, email string) (int, bool, error) {
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
