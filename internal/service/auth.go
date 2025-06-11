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
	Verify(ctx context.Context, code string) error
	RefreshToken(ctx context.Context, refreshToken string) (models.Tokens, error)
	Logout(ctx context.Context, refreshToken string) error
}

type authService struct {
	repo         repository.Auth
	log          *slog.Logger
	cfg          *config.Config
	emailService mail.Sender
	jwt          jwt.JWT
}

func NewAuthService(repo repository.Auth, log *slog.Logger, cfg *config.Config, emailService mail.Sender, jwt jwt.JWT) *authService {
	return &authService{
		repo:         repo,
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

	authData.Password = hash.PasswordHash(authData.Password)
	id, verified, err := a.repo.GetByLoginAndPassword(ctx, authData)
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

	userID, err := a.repo.CreateUser(ctx, &newUser)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return a.handleExistingUser(ctx, authData)
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return a.triggerVerification(ctx, userID, authData.Email)
}

func (a *authService) Verify(ctx context.Context, code string) error {
	const op = "service.Verify"

	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	err := a.repo.Verify(ctx, code)
	return a.handleRepositoryError(op, err, repository.ErrUserNotFound)
}

func (a *authService) RefreshToken(ctx context.Context, refreshToken string) (models.Tokens, error) {
	const op = "service.RefreshToken"

	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	refreshTokenData, err := a.ValidateRefreshToken(ctx, refreshToken)
	if err != nil {
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}
	return a.createSession(ctx, refreshTokenData.UserID, refreshTokenData.DeviceID)
}

func (a *authService) Logout(ctx context.Context, refreshToken string) error {
	const op = "service.Logout"

	ctx, cancel := a.newRequestContext(ctx)
	defer cancel()

	hashToken := hash.Token(refreshToken)
	err := a.repo.DeleteRefreshToken(ctx, hashToken)
	return a.handleRepositoryError(op, err, repository.ErrUserNotFound)
}

func (a *authService) createSession(ctx context.Context, userID int, deviceID string) (models.Tokens, error) {
	const op = "service.createSession"

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

	if err := a.repo.CreateOrUpdateRefreshToken(ctx, userID, deviceID, hashToken, expiresAt); err != nil {
		return models.Tokens{}, a.logAndWrapError(op, "failed to save refresh token", err)
	}

	return models.Tokens{
		AccessToken:  tokenAccess,
		RefreshToken: tokenRefresh,
	}, nil
}

func (a *authService) triggerVerification(ctx context.Context, userID int, email string) error {
	code, err := a.initiateVerification(ctx, userID)
	if err != nil {
		return err
	}
	return a.sendVerification(code, email)
}

func (a *authService) initiateVerification(ctx context.Context, userID int) (string, error) {
	const op = "service.initiateVerification"
	code, err := auth.GenerateMailCode()
	if err != nil {
		return "", a.logAndWrapError(op, "failed to generate mail code", err)
	}

	expiresAt := time.Now().Add(a.cfg.Email.ExpiresAt)
	if err := a.repo.CreateVerificationCode(ctx, userID, code, expiresAt); err != nil {
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

func (a *authService) handleExistingUser(ctx context.Context, authData *models.FirstAuth) error {
	const op = "service.handleExistingUser"

	id, verified, err := a.repo.GetUserByEmail(ctx, authData.Email)
	if err != nil {
		return fmt.Errorf("%s: failed to get existing user: %w", op, err)
	}

	if verified {
		return repository.ErrUserExists
	}
	a.log.Info("re-triggering verification for unverified user", "userID", id)
	return a.triggerVerification(ctx, id, authData.Email)
}

func (a *authService) ValidateRefreshToken(ctx context.Context, refreshToken string) (*models.RefreshTokenData, error) {
	const op = "service.ValidateRefreshToken"

	hashToken := hash.Token(refreshToken)
	refreshTokenData, err := a.repo.GetRefreshTokenData(ctx, hashToken)
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
