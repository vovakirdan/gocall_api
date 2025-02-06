package utils

import (
	"os"
	"time"
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey []byte

func GenerateJWT(userID int) (string, error) {
	jwtKey = []byte(os.Getenv("SECRET_KEY"))
	claims := &jwt.MapClaims{
		"user_id": userID,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jwtKey)
}

func DecodeJWT(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
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
