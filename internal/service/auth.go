package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"todoai/internal/config"
	"todoai/internal/models"
	"todoai/internal/repository"
	"todoai/pkg/auth"
	"todoai/pkg/hash"
	"todoai/pkg/jwt"
	"todoai/pkg/mail"
)

var (
	ErrUserNotVerified = errors.New("user not verified")
	ErrTokenExpired    = errors.New("token expired")
)

//go:generate mockgen -source=auth.go -destination=mocks/auth_mock.go -package=mocks

type AuthService interface {
	SignUp(ctx context.Context, authData *models.FirstAuth) error
	SignIn(ctx context.Context, authData *models.SecondAuth) (models.Tokens, error)
	VerifyUser(ctx context.Context, code string) error
	RefreshToken(ctx context.Context, refreshToken string) (models.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
	ResendCode(ctx context.Context, email string) error
}

type authService struct {
	tm           *TransactionManager
	log          *slog.Logger
	cfg          *config.Config
	emailService mail.Sender
	jwt          jwt.JWT
}

func NewAuthService(tm *TransactionManager, log *slog.Logger, cfg *config.Config, emailService mail.Sender, jwt jwt.JWT) *authService {
	return &authService{
		tm:           tm,
		log:          log,
		cfg:          cfg,
		emailService: emailService,
		jwt:          jwt,
	}
}

func (a *authService) SignIn(ctx context.Context, authData *models.SecondAuth) (models.Tokens, error) {
	const op = "service.SignIn"

	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	repo := a.tm.NewAuthRepo()
	authData.Password = hash.PasswordHash(authData.Password)
	id, verified, err := repo.GetByLoginAndPassword(ctx, authData)
	if err != nil {
		return models.Tokens{}, a.handleRepositoryError(op, err, repository.ErrUserNotFound)
	}

	if !verified {
		return models.Tokens{}, fmt.Errorf("%s: %w", op, ErrUserNotVerified)
	}

	return a.createSession(ctx, id, authData.DeviceID)
}

func (a *authService) SignUp(ctx context.Context, authData *models.FirstAuth) error {
	const op = "service.SignUp"

	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	newUser := models.FirstAuth{
		Email:    authData.Email,
		Login:    authData.Login,
		Password: hash.PasswordHash(authData.Password),
		Verified: false,
	}

	var verificationCode string

	err := a.tm.WithTransaction(ctx, func(repos *TransactionalRepos) error {
		userID, err := repos.Auth.CreateUser(ctx, &newUser)
		if err != nil {
			return err
		}

		code, err := a.initiateVerification(ctx, repos.Auth, userID)
		if err != nil {
			return err
		}
		verificationCode = code
		return nil
	})

	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return fmt.Errorf("failed to set user verified: %w", repository.ErrUserExists)
		}
		a.log.Error("transaction failed during sign up", "error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return a.sendVerification(verificationCode, authData.Email)
}

func (a *authService) VerifyUser(ctx context.Context, code string) error {
	const op = "service.VerifyUser"

	err := a.tm.WithTransaction(ctx, func(repo *TransactionalRepos) error {
		userID, err := repo.Auth.GetUserIDByVerificationCode(ctx, code)
		if err != nil {
			return err
		}

		if err := repo.Auth.SetUserVerified(ctx, userID); err != nil {
			a.log.Error(op, "failed to set user verified: %w", err)
			return fmt.Errorf("failed to set user verified: %w", err)
		}
		if err := repo.Auth.DeleteVerificationCode(ctx, code); err != nil {
			a.log.Error("failed to delete verification code", "code", code, "error", err)
			return fmt.Errorf("failed to delete verification code: %w", err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *authService) ResendCode(ctx context.Context, email string) error {
	return a.handleExistingUser(ctx, email)
}

func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (models.Tokens, error) {
	const op = "service.RefreshToken"

	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	refreshTokenData, err := a.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if time.Until(refreshTokenData.ExpiresAt) > 24*time.Hour {
		accessToken, err := a.jwt.GenerateAccessToken(refreshTokenData.UserID)
		if err != nil {
			return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
		}
		return models.Tokens{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}, nil
	}

	return a.createSession(ctx, refreshTokenData.UserID, refreshTokenData.DeviceID)
}

func (a *authService) Logout(ctx context.Context, refreshToken string) error {
	const op = "service.Logout"
	repo := a.tm.NewAuthRepo()
	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	hashToken := hash.Token(refreshToken)
	err := repo.DeleteRefreshToken(ctx, hashToken)
	return a.handleRepositoryError(op, err, repository.ErrUserNotFound)
}

func (a *authService) createSession(ctx context.Context, userID int, deviceID string) (models.Tokens, error) {
	const op = "service.createSession"
	repo := a.tm.NewAuthRepo()

	tokenAccess, err := a.jwt.GenerateAccessToken(userID)
	if err != nil {
		return models.Tokens{}, a.logAndWrapError(op, "failed to generate access token", err)
	}

	tokenRefresh, err := a.jwt.GenerateRefreshToken(userID, a.cfg.Session.ExpiresAt)
	if err != nil {
		return models.Tokens{}, a.logAndWrapError(op, "failed to generate refresh token", err)
	}

	hashToken := hash.Token(tokenRefresh)
	expiresAt := time.Now().Add(a.cfg.Session.ExpiresAt)

	if err := repo.CreateOrUpdateRefreshToken(ctx, userID, deviceID, hashToken, expiresAt); err != nil {
		return models.Tokens{}, a.logAndWrapError(op, "failed to save refresh token", err)
	}

	return models.Tokens{
		AccessToken:  tokenAccess,
		RefreshToken: tokenRefresh,
	}, nil
}

func (a *authService) triggerVerification(ctx context.Context, userID int, email string) error {
	repo := a.tm.NewAuthRepo()
	code, err := a.initiateVerification(ctx, repo, userID)
	if err != nil {
		return err
	}

	return a.sendVerification(code, email)
}

func (a *authService) initiateVerification(ctx context.Context, repo repository.Auth, userID int) (string, error) {
	const op = "service.initiateVerification"

	code, err := auth.GenerateMailCode()
	if err != nil {
		return "", a.logAndWrapError(op, "failed to generate mail code", err)
	}

	expiresAt := time.Now().Add(a.cfg.Email.ExpiresAt)
	if err := repo.UpsertVerificationCode(ctx, userID, code, expiresAt); err != nil {
		return "", a.logAndWrapError(op, "failed to create verification code", err)
	}
	return code, nil
}

func (a *authService) sendVerification(code, email string) error {
	const op = "service.sendVerification"

	mailData := &mail.MailData{
		To:      email,
		Subject: "Verification email",
		Body:    fmt.Sprintf("Verification code: %s", code),
	}
	if err := a.emailService.Send(mailData); err != nil {
		return a.logAndWrapError(op, "failed to send verification email", err)
	}

	return nil
}

func (a *authService) handleExistingUser(ctx context.Context, email string) error {
	const op = "service.handleExistingUser"

	repo := a.tm.NewAuthRepo()
	id, verified, err := repo.GetUserByEmail(ctx, email)
	if err != nil {
		return a.handleRepositoryError(op, err, repository.ErrUserNotFound)
	}

	if verified {
		return repository.ErrUserExists
	}

	return a.triggerVerification(ctx, id, email)
}

func (a *authService) ValidateRefreshToken(ctx context.Context, refreshToken string) (*models.RefreshTokenData, error) {
	const op = "service.ValidateRefreshToken"
	repo := a.tm.NewAuthRepo()
	hashToken := hash.Token(refreshToken)
	refreshTokenData, err := repo.GetRefreshTokenData(ctx, hashToken)
	if err != nil {
		return nil, a.handleRepositoryError(op, err, repository.ErrUserNotFound)
	}

	if refreshTokenData.ExpiresAt.Before(time.Now()) {
		return nil, ErrTokenExpired
	}

	return &refreshTokenData, nil
}

func (a *authService) newRequestContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, a.cfg.Database.Timeout)
}

func (a *authService) handleRepositoryError(op string, err error, specificUserErr error) error {
	if errors.Is(err, specificUserErr) {
		a.log.Warn(op, slog.String("error", err.Error()))
		return fmt.Errorf("%s: %w", op, specificUserErr)
	}
	return a.logAndWrapError(op, "repository error", err)
}

func (a *authService) logAndWrapError(op, msg string, err error) error {
	a.log.Error(op, slog.String("message", msg), slog.String("error", err.Error()))
	return fmt.Errorf("%s: %w", op, err)
}
