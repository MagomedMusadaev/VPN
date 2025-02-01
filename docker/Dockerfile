# Используем официальный образ Go
FROM golang:1.21

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем зависимости
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Устанавливаем переменные окружения
ENV GO_ENV=production

# Собираем бинарник
RUN go build -o main ./cmd/main.go

# Запускаем бота
CMD ["./main"]
