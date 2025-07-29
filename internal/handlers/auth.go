package handlers

import (
	"auth-service/internal/api_clients"
	"auth-service/internal/client"
	"auth-service/internal/dtos"
	"auth-service/internal/services"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type handlers struct {
	cfg             *dtos.Config
	dbServiceClient *api_clients.DbServiceClient
}

func Register(app *fiber.App) {
	h := &handlers{
		cfg:             services.NewConfig(),
		dbServiceClient: api_clients.NewDbServiceClient(),
	}

	app.Post("/register", h.registerHandler)
	app.Post("/login", h.loginHandler)
	app.Get("/validate", h.validateTokenHandler)
}

func (h *handlers) registerHandler(c fiber.Ctx) error {
	var req dtos.RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	registeredUser, err := h.dbServiceClient.RegisterUser(&req)
	if err != nil {
		slog.Error("Failed to register user", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(registeredUser)
}

func (h *handlers) loginHandler(c fiber.Ctx) error {
	var req dtos.LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Отправляем запрос в db-service
	resp, err := client.Post("/user/login", req)
	if err != nil {
		slog.Error("Failed to connect to db-service", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Service unavailable"})
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).Send(body)
	}

	var user dtos.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		slog.Error("Failed to parse user data", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.RoleID,
		"exp":  time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		slog.Error("Token generation failed", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Token generation failed"})
	}

	return c.JSON(fiber.Map{
		"token": tokenString,
		"user":  user,
	})
}

func (h *handlers) validateTokenHandler(c fiber.Ctx) error {
	tokenString := c.Query("token")
	if tokenString == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Token is required"})
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	return c.JSON(fiber.Map{
		"valid":    true,
		"userID":   claims["sub"],
		"userRole": claims["role"],
	})
}
