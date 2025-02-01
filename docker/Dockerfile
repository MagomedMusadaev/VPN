# Используем официальный образ Go
FROM golang:1.21

# Устанавливаем рабочую директорию в контейнере
WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем весь проект для сборки
COPY . .

# Устанавливаем переменные окружения
ENV GO_ENV=production

# Собираем бинарный файл
RUN go build -o main ./cmd/main.go

# Запуск бота
CMD ["./main"]
