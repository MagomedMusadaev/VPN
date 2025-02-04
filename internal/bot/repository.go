package bot

import (
	"bot_vpn/internal/entities"
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

// IsUserInDB - функция проверяет наличие пользователя в базе данных по userID.
func (r *Repo) IsUserInDB(userID int) (bool, error) {
	const op = "internal/bot/repository.go/IsUserInDB"

	// SQL запрос для проверки наличия пользователя в базе данных по userID.
	query := "SELECT COUNT(1) FROM vpn WHERE user_id = $1"

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		// Если ошибка — проверяем, относится ли она к отсутствию строк.
		if err == sql.ErrNoRows {
			slog.Warn("Пользователь не найден в базе", slog.Int("userID", userID))
			return false, nil
		}
		// Логируем ошибку выполнения запроса.
		slog.Error(op, "Ошибка при проверке пользователя в базе", slog.Int("userID", userID), slog.String("error", err.Error()))
		return false, err
	}

	// Логируем успешную проверку
	slog.Info("Проверка пользователя в базе", slog.Int("userID", userID), slog.Bool("exists", count > 0))

	// Если count > 0, значит пользователь существует в базе.
	return count > 0, nil
}

// CreateUser - функция записи нового пользователя в db.
func (r *Repo) CreateUser(user *entities.User) error {
	const op = "internal/bot/repository.go/CreateUser"

	query := `INSERT INTO users (telegram_id, chat_id, referral_code, created_at, referred_by) VALUES($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, user.UserTgID, user.ChatTgID, user.ReferralCode, user.CreatedAt, user.ReferredBy)
	if err != nil {
		slog.Error(op, slog.String("error", err.Error()))
		return err
	}

	return nil
}

// SaveUserKey - функция записи нового ключа в db.
func (r *Repo) SaveUserKey(key *entities.Key) error {
	const op = "internal/bot/repository.go/SaveUserKey"

	query := `INSERT INTO keys (user_id, key, created_at, expires_at) VALUES($1, $2, $3, $4)`

	_, err := r.db.Exec(query, key.UserTgID, key.Key, key.CreatedAt, key.ExpiresAt)
	if err != nil {
		slog.Error(op, slog.String("error", err.Error()))
		return err
	}

	return nil
}

// GetExpirationTimeKey - функция получения времени истечения ключа пользователя.
func (r *Repo) GetExpirationTimeKey(userTgID int) (time.Time, error) {
	const op = "internal/bot/repository.go/GetExpirationTimeKey"

	var expiresAt time.Time
	query := `SELECT expires_at FROM keys WHERE user_id = $1`

	err := r.db.QueryRow(query, userTgID).Scan(&expiresAt)
	if err != nil {
		slog.Error(op, slog.String("error", err.Error()))
		return time.Time{}, err
	}

	return expiresAt, nil
}

func (r *Repo) UpdateKeyExpiration(userTgID int, newExpiration time.Time) error {
	const op = "internal/bot/repository.go/UpdateKeyExpiration"

	query := `INSERT INTO keys (expires_at) VALUES($1) WHERE user_id = $2 `

	_, err := r.db.Exec(query, userTgID, newExpiration)
	if err != nil {
		slog.Error(op, slog.String("error", err.Error()))
		return err
	}

	return nil
}

//func (r *Repo) UpdateKeyExpiration(userID int64) error {
//	query := "UPDATE vpn SET key_expiration_date = NOW() + (добавить как-то время для ключа по выбранному тарифу) WHERE user_tg_id = $1"
//	_, err := r.db.Exec(query, userID)
//	return err
//}

// AddUserToRedis - добавляет пользователя в Redis на временное хранение.
//func (r *Repo) AddUserToRedis(userID, chatID string, ttl time.Duration) error {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
//	defer cancel()
//
//	// Попытка установить значение в Redis по ключу userID с временем жизни ttl.
//	err := r.client.Set(ctx, userID, chatID, ttl).Err()
//	if err != nil {
//		slog.Error("Ошибка при записи данных в Redis", slog.String("error", err.Error()))
//		return err
//	}
//	slog.Info("Пользователь успешно добавлен в Redis до оплаты:", slog.String("userID", userID))
//
//	return nil
//}

// CheckRedisUserData - проверяет наличие пользователя в Redis.
//func (r *Repo) CheckRedisUserData(userTgID string) (string, error) {
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
//	defer cancel()
//
//	//
//	chatID, err := r.client.Get(ctx, userTgID).Result()
//	if err == redis.Nil {
//		slog.Warn("Данные в Redis не найдены", slog.String("key", userTgID))
//		return "", nil // Отсутствие данных — не ошибка
//	}
//	if err != nil {
//		slog.Error("Ошибка при  данных в Redis", slog.String("error", err.Error()))
//		return "", err
//	}
//
//	return chatID, nil
//}

//// GetUserFromRedis получает значение chatID по ключу userID из Redis.
//func (r *Repo) GetUserFromRedis(userID string) (string, error) {
//	const op = "internal/bot/repository.go/GetUserFromRedis"
//
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
//	defer cancel()
//
//	// Попытка получить значение chatID из Redis по ключу userID.
//	chatID, err := r.client.Get(ctx, userID).Result()
//	if err != nil {
//		// Если ошибка Redis.Nil, то ключ не существует в Redis, возвращаем пустую строку и nil.
//		if err == redis.Nil {
//			return "", nil
//		}
//		slog.Error(op, "Ошибка при чтении данных из Redis", slog.String("error", err.Error()))
//		return "", err
//	}
//
//	// Логируем успешное получение данных для пользователя.
//	slog.Info("Пользователь успешно найден в Redis", slog.String("userID", userID))
//
//	return chatID, nil
//}
