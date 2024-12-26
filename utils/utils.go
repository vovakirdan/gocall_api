package utils

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

const (
	SECRET_KEY = "SECRET_KEY"
	ALLOW_ORIGINS = "ALLOW_ORIGINS"
)

func CheckEnvLoaded() error {
	godotenv.Load()
	if os.Getenv(SECRET_KEY) == "" {
		return errors.New("SECRET_KEY not found")
	}

	if os.Getenv(ALLOW_ORIGINS) == "" {
		return errors.New("ALLOW_ORIGINS not found")
	}
	
	return nil
}