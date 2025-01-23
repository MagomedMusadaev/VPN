package bot

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"strconv"
)

// NewRedis - создает и возвращает клиента для подключения к Redis.
func NewRedis() *redis.Client {
	const op = "internal/bot/connectRepo/NewRedis"

	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASS")
	redisDB := os.Getenv("REDIS_DB")

	intRedisDB, _ := strconv.Atoi(redisDB) // Конвертация номера базы в целое число

	// Создание клиента Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,     // Адрес Redis
		Password: redisPassword, // Пароль Redis
		DB:       intRedisDB,    // Номер базы данных Redis
	})

	// Проверка подключения
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error(op, "Не удалось подключиться к Redis:", err)
		return nil
	}

	slog.Info("Подключение к Redis успешно установлено")
	return rdb
}

// ConnectPostgresDB - подключается к базе данных PostgreSQL и возвращает объект соединения.
func ConnectPostgresDB() *sql.DB {
	const op = "internal/bot/connectRepo/ConnectPostgresDB"

	hostPSQL := os.Getenv("POSTGRES_HOST")
	portPSQL := os.Getenv("POSTGRES_PORT")
	userPSQL := os.Getenv("POSTGRES_USER")
	passPSQL := os.Getenv("POSTGRES_PASS")
	dbnamePSQL := os.Getenv("POSTGRES_NAME")

	// Формирование строки подключения
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		hostPSQL, portPSQL, userPSQL, passPSQL, dbnamePSQL,
	)

	// Подключение к базе данных
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		slog.Error(op, "Не удалось подключиться к db:", err)
		return nil
	}

	// Проверка соединения
	if err := db.Ping(); err != nil {
		slog.Error(op, "Нет коннекта с db:", err)
		return nil
	}

	slog.Info("Успешно подключено к PostgresDB")

	return db
}
