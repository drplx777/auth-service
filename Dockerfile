# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Установка зависимостей
RUN apk add --no-cache git

# Установка рабочей директории
WORKDIR /app

# Копируем файлы модулей и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service

# Stage 2: Create the final image
FROM alpine:latest

# Установка зависимостей времени выполнения
RUN apk --no-cache add ca-certificates

# Копируем бинарник из builder
COPY --from=builder /auth-service /auth-service

# Открываем порт сервиса
EXPOSE 5000

# Команда для запуска сервиса
CMD ["/auth-service"]