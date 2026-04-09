package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func signingKey() []byte {
	return []byte(os.Getenv(SECRET_KEY))
}

// GenerateJWT creates a signed JWT for the given numeric user ID.
func GenerateJWT(userID int) (string, error) {
	claims := &jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(signingKey())
}

// DecodeJWT parses a token string and returns the embedded numeric user ID.
func DecodeJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey(), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	userID, ok := claims["user_id"].(float64)
	if !ok {
		return 0, errors.New("user_id is not a float64")
	}

	return int(userID), nil
}
