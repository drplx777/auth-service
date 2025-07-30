# Stage 1: Build the application
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd/server/main.go

# Stage 2: Create the final image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Устанавливаем рабочую директорию, где лежит go.mod
WORKDIR /app

# Копируем весь /app из билдера (включая go.mod, .env и бинарь)
COPY --from=builder /app .

EXPOSE 5000

# Запускаем сервис из /app, где есть go.mod
CMD ["./auth-service"]
