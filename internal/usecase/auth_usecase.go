package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/yourorg/boardgov/config"
	"github.com/yourorg/boardgov/internal/domain"
	"github.com/yourorg/boardgov/pkg/crypto"
	"github.com/yourorg/boardgov/pkg/logger"
)

type AuthUsecase interface {
	Register(ctx context.Context, email, password, fullName string) (*domain.User, error)
	Login(ctx context.Context, email, password, ip string) (*TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)
	Logout(ctx context.Context, refreshToken string) error
	ValidateAccessToken(ctx context.Context, tokenStr string) (*Claims, error)
}

type TokenPair struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    time.Time    `json:"expires_at"`
	User         *domain.User `json:"user,omitempty"`
}

type Claims struct {
	UserID      uuid.UUID `json:"user_id"`
	Email       string    `json:"email"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

type authUsecase struct {
	userRepo  domain.UserRepository
	tokenRepo domain.RefreshTokenRepository
	auditRepo domain.AuditLogRepository
	cfg       config.JWTConfig
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	tokenRepo domain.RefreshTokenRepository,
	auditRepo domain.AuditLogRepository,
	cfg config.JWTConfig,
) AuthUsecase {
	return &authUsecase{userRepo: userRepo, tokenRepo: tokenRepo, auditRepo: auditRepo, cfg: cfg}
}

func (u *authUsecase) Register(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	hash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	user := &domain.User{
		ID:           uuid.New(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: hash,
		IsActive:     true,
	}
	if err := u.userRepo.Create(ctx, user); err != nil {
		logger.Log.Info().Msg("error creating user: " + err.Error())
		return nil, err
	}
	return user, nil
}

func (u *authUsecase) Login(ctx context.Context, email, password, ip string) (*TokenPair, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrUnauthorized
		}
		return nil, err
	}

	if !user.IsActive {
		return nil, domain.ErrUnauthorized
	}

	if !crypto.CheckPassword(user.PasswordHash, password) {
		return nil, domain.ErrUnauthorized
	}

	perms, err := u.userRepo.ListPermissions(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}

	pair, err := u.generateTokenPair(ctx, user, perms)
	if err != nil {
		return nil, err
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	_ = u.userRepo.Update(ctx, user)

	// Audit
	u.logAudit(ctx, user.ID, "user.login", ip, nil, nil, "", "")

	logger.Log.Info().Str("user_id", user.ID.String()).Str("email", email).Msg("user logged in")
	return pair, nil
}

func (u *authUsecase) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	hash := crypto.SHA256String(refreshToken)
	stored, err := u.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return nil, domain.ErrUnauthorized
	}

	// Revoke old token
	_ = u.tokenRepo.Revoke(ctx, stored.ID)

	user, err := u.userRepo.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	perms, err := u.userRepo.ListPermissions(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	return u.generateTokenPair(ctx, user, perms)
}

func (u *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	hash := crypto.SHA256String(refreshToken)
	stored, err := u.tokenRepo.GetByHash(ctx, hash)
	if err != nil {
		return nil // already gone
	}
	return u.tokenRepo.Revoke(ctx, stored.ID)
}

func (u *authUsecase) ValidateAccessToken(ctx context.Context, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(u.cfg.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

func (u *authUsecase) generateTokenPair(ctx context.Context, user *domain.User, perms []string) (*TokenPair, error) {
	now := time.Now()
	accessExpiry := now.Add(u.cfg.AccessExpiry)

	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Permissions: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    u.cfg.Issuer,
			Subject:   user.ID.String(),
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(u.cfg.AccessSecret))
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	rawRefresh, err := crypto.RandomHex(32)
	if err != nil {
		return nil, err
	}

	rt := &domain.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: crypto.SHA256String(rawRefresh),
		ExpiresAt: now.Add(u.cfg.RefreshExpiry),
		User:      user,
	}
	if err := u.tokenRepo.Create(ctx, rt); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    accessExpiry,
		User:         user,
	}, nil
}

func (u *authUsecase) logAudit(ctx context.Context, userID uuid.UUID, action, ip string, meetingID, docID *uuid.UUID, prev, next string) {
	entry := &domain.AuditLog{
		UserID:        &userID,
		Action:        action,
		MeetingID:     meetingID,
		DocumentID:    docID,
		PreviousState: prev,
		NewState:      next,
		IPAddress:     ip,
	}
	if err := u.auditRepo.Create(ctx, entry); err != nil {
		logger.Log.Error().Err(err).Msg("failed to write audit log")
	}
}
