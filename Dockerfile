# Dockerfile
FROM golang:1.24-alpine

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Сборка бинарника
RUN go build -o main cmd/api/main.go

# Запуск
CMD ["./main"]