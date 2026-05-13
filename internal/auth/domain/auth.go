package domain

import (
	"context"
)

type AuthUsecase interface {
	Login(ctx context.Context, username, password string) (accessToken string, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (accessToken string, newRefreshToken string, err error)
}
