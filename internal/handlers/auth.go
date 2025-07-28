package handler

import (
	"auth-service/internal/client"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

type RegisterRequest struct {
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename,omitempty"`
	Login      string `json:"login"`
	RoleID     int    `json:"roleID"`
	Password   string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Surname    string `json:"surname"`
	Middlename string `json:"middlename,omitempty"`
	Login      string `json:"login"`
	RoleID     int    `json:"roleID"`
}

func RegisterAuthRoutes(app *fiber.App) {
	app.Post("/register", registerHandler)
	app.Post("/login", loginHandler)
	app.Get("/validate", validateTokenHandler)
}

func registerHandler(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Отправляем запрос в db-service
	resp, err := client.Post("/user/register", req)
	if err != nil {
		slog.Error("Failed to connect to db-service", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Service unavailable"})
	}
	defer resp.Body.Close()

	// Если db-service вернул ошибку, передаем ее клиенту
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).Send(body)
	}

	// Парсим ответ от db-service
	var userID struct {
		ID int `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userID); err != nil {
		slog.Error("Failed to parse response", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(userID)
}

func loginHandler(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Отправляем запрос в db-service
	resp, err := client.Post("/user/login", req)
	if err != nil {
		slog.Error("Failed to connect to db-service", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Service unavailable"})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return c.Status(resp.StatusCode).Send(body)
	}

	var user UserResponse
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

func validateTokenHandler(c fiber.Ctx) error {
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
