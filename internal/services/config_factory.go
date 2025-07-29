package services

import (
	"auth-service/internal/dtos"
	"github.com/KoNekoD/dotenv/pkg/dotenv"
	"github.com/pkg/errors"
	"os"
)

func NewConfig() *dtos.Config {
	err := dotenv.LoadEnv(".env")
	if err != nil {
		panic(errors.Wrap(err, "failed to load .env file"))
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic(errors.New("JWT_SECRET environment variable is required"))
	}

	return &dtos.Config{
		JwtSecret: jwtSecret,
	}
}
