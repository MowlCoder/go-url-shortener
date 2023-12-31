package jwt

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const (
	defaultJWTSecret = "secret"
	envKeyJWTSecret  = "JWT_SECRET"
)

// Claims is JWT payload.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// GenerateToken generate JWT token with payload containing given user id.
func GenerateToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(time.Hour * 24)),
		},
		UserID: userID,
	})

	tokenString, err := token.SignedString([]byte(getJWTSecretKey()))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken parsing JWT token and return JWT payload.
func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(getJWTSecretKey()), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}

func getJWTSecretKey() string {
	key, ok := os.LookupEnv(envKeyJWTSecret)

	if !ok {
		return defaultJWTSecret
	}

	return key
}
