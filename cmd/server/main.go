package main

import (
	"auth-service/internal/handlers"
	"log"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://my-samovar-to-do-list.duckdns.org/", "http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		ExposeHeaders:    []string{"Authorization"},
	}))

	handlers.Register(app)

	slog.Info("Starting auth-service", "port", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start auth-service: %v", err)
	}
}
