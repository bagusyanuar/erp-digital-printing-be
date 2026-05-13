package usecase

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/bagusyanuar/erp-digital-printing-be/internal/auth/domain"
	userDomain "github.com/bagusyanuar/erp-digital-printing-be/internal/user/domain"
	"github.com/bagusyanuar/erp-digital-printing-be/internal/shared/config"
	"github.com/bagusyanuar/erp-digital-printing-be/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"go.uber.org/zap"
)

type authUsecase struct {
	userRepo   userDomain.UserRepository
	accessJWT  jwt.JWTUtil
	refreshJWT jwt.JWTUtil
	cfg        *config.Config
	logger     *zap.Logger
}

func NewAuthUsecase(
	userRepo userDomain.UserRepository,
	accessJWT jwt.JWTUtil,
	refreshJWT jwt.JWTUtil,
	cfg *config.Config,
	logger *zap.Logger,
) domain.AuthUsecase {
	return &authUsecase{
		userRepo:   userRepo,
		accessJWT:  accessJWT,
		refreshJWT: refreshJWT,
		cfg:        cfg,
		logger:     logger,
	}
}

func (u *authUsecase) Login(ctx context.Context, username, password string) (string, string, error) {
	user, err := u.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", "", errors.New("invalid username or password")
	}

	// Extract Roles & Permissions
	roles, perms := u.extractRolesAndPermissions(user)

	// Generate Access Token
	accessToken, err := u.accessJWT.GenerateToken(
		user.ID,
		user.Username,
		roles,
		perms,
		time.Duration(u.cfg.JWT.Expiration)*time.Minute,
	)
	if err != nil {
		return "", "", err
	}

	// Generate Refresh Token
	refreshToken, err := u.refreshJWT.GenerateToken(
		user.ID,
		user.Username,
		roles,
		perms,
		time.Duration(u.cfg.JWT.ExpirationRefresh)*time.Hour*24,
	)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (u *authUsecase) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// 1. Parse & Validate Refresh Token
	claims, err := u.refreshJWT.ParseToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// 2. Fetch User to get latest roles/permissions
	user, err := u.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return "", "", errors.New("user not found")
	}

	// 3. Extract Roles & Permissions
	roles, perms := u.extractRolesAndPermissions(user)

	// 4. Generate New Tokens (Token Rotation)
	newAccessToken, err := u.accessJWT.GenerateToken(
		user.ID,
		user.Username,
		roles,
		perms,
		time.Duration(u.cfg.JWT.Expiration)*time.Minute,
	)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err := u.refreshJWT.GenerateToken(
		user.ID,
		user.Username,
		roles,
		perms,
		time.Duration(u.cfg.JWT.ExpirationRefresh)*time.Hour*24,
	)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}

func (u *authUsecase) extractRolesAndPermissions(user *userDomain.User) ([]string, []string) {
	roles := make([]string, 0, len(user.Roles))
	permsMap := make(map[string]bool)
	isAdmin := false

	for _, r := range user.Roles {
		roles = append(roles, r.Name)
		if r.Name == "administrator" {
			isAdmin = true
		}
		for _, p := range r.Permissions {
			permsMap[p.Resource+":"+p.Action] = true
		}
	}

	perms := make([]string, 0, len(permsMap))
	if isAdmin {
		perms = append(perms, "*:*")
	} else {
		for p := range permsMap {
			perms = append(perms, p)
		}
		sort.Strings(perms)
	}

	return roles, perms
}
