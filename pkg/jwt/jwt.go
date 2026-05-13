package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	Username    string    `json:"username"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	jwt.RegisteredClaims
}

type JWTUtil interface {
	GenerateToken(userID uuid.UUID, username string, roles []string, perms []string, duration time.Duration) (string, error)
	ParseToken(tokenString string) (*Claims, error)
}

type jwtUtil struct {
	secretKey []byte
	issuer    string
}

func NewJWTUtil(secretKey string, issuer string) JWTUtil {
	return &jwtUtil{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

func (j *jwtUtil) GenerateToken(userID uuid.UUID, username string, roles []string, perms []string, duration time.Duration) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Username:    username,
		Roles:       roles,
		Permissions: perms,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *jwtUtil) ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
