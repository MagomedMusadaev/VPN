package bot

import (
	"context"
	"database/sql"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"time"
)

type RepoInt interface {
	IsUserInDB(userID int64) (bool, error)
}

type Repo struct {
	db     *sql.DB
	client *redis.Client
}

func NewRepo(db *sql.DB, client *redis.Client) *Repo {
	return &Repo{
		db:     db,
		client: client,
	}
}

// IsUserInDB - Функция проверяет наличие пользователя в базе данных по userID.
func (r *Repo) IsUserInDB(userID int64) (bool, error) {
	// SQL запрос для проверки наличия пользователя в базе данных по userID.
	query := "SELECT COUNT(1) FROM users WHERE user_id = $1"

	var count int
	// Выполняем запрос и сканируем результат в переменную count.
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		// Если ошибка — проверяем, относится ли она к отсутствию строк.
		//if err == sql.ErrNoRows {
		//	return false, nil
		//}
		// Логируем ошибку и возвращаем её.
		slog.Error("internal/bot/repository/IsUserInDB", err)
		return false, nil
	}

	// Если count > 0, значит пользователь существует в базе.
	return count > 0, nil
}

// AddUserToRedis - добавляет пользователя в Redis на временное хранение.
func (r *Repo) AddUserToRedis(userID, chatID string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Попытка установить значение в Redis по ключу userID с временем жизни ttl.
	err := r.client.Set(ctx, userID, chatID, ttl).Err()
	if err != nil {
		slog.Error("Ошибка при записи данных в Redis", slog.String("error", err.Error()))
		return err
	}
	slog.Info("Пользователь успешно добавлен в Redis на временное хранение", slog.String("userID", userID))

	return nil
}
