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
	query := "SELECT COUNT(1) FROM vpn WHERE user_id = $1"

	var count int
	// Выполняем запрос и сканируем результат в переменную count.
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		// Если ошибка — проверяем, относится ли она к отсутствию строк.
		//if err == sql.ErrNoRows {
		//	return false, nil
		//}
		// Логируем ошибку и возвращаем её.
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
	slog.Info("Пользователь успешно добавлен в Redis до оплаты:", slog.String("userID", userID))

	return nil
}

func (r *Repo) UpdateKeyExpiration(userID int64) error {
	query := "UPDATE vpn SET key_expiration_date = NOW() + (добавить как-то время для ключа по выбранному тарифу) WHERE user_tg_id = $1"
	_, err := r.db.Exec(query, userID)
	return err
}

// GetUserFromRedis получает значение chatID по ключу userID из Redis.
func (r *Repo) GetUserFromRedis(userID string) (string, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Попытка получить значение chatID из Redis по ключу userID.
	chatID, err := r.client.Get(ctx, userID).Result()
	if err != nil {
		// Если ошибка Redis.Nil, то ключ не существует в Redis, возвращаем пустую строку и nil.
		if err == redis.Nil {
			return "", nil
		}
		slog.Error("Ошибка при чтении данных из Redis", slog.String("error", err.Error()))
		return "", err
	}

	// Логируем успешное получение данных для пользователя.
	slog.Info("Пользователь успешно найден в Redis", slog.String("userID", userID))

	return chatID, nil
}
