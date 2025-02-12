package bot

import (
	"bot_vpn/internal/entities"
	"database/sql"
	"log/slog"
	"time"
)

type RepoInt interface {
	IsUserInDB(userID int) (bool, error)
	CreateUser(user *entities.User) error
	SaveUserKey(key *entities.Key) error
	GetExpirationTimeKey(userTgID int) (time.Time, error)
	UpdateKeyExpiration(userTgID int, newExpiration time.Time) error
	GetConnectionKey(userTgID int) (string, error)
	GetConnectKeyOrReferral(userTgID int64, referral bool) (string, error)
}

//type Repo struct {
//	db     *sql.DB
//	client *redis.Client
//}

//func NewRepo(db *sql.DB, client *redis.Client) *Repo {
//	return &Repo{
//		db:     db,
//		client: client,
//	}
//}

type Repo struct {
	db *sql.DB
}

func NewRepo(db *sql.DB) *Repo {
	return &Repo{
		db: db,
	}
}

// IsUserInDB - функция проверяет наличие пользователя в базе данных по userID.
func (r *Repo) IsUserInDB(userID int) (bool, error) {
	const op = "internal/bot/repository.go/IsUserInDB"

	query := "SELECT COUNT(1) FROM users WHERE telegram_id = $1"

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		// Если ошибка — проверяем, относится ли она к отсутствию строк.
		if err == sql.ErrNoRows {
			// Пользователь не найден, логируем это событие.
			slog.Warn(op, "Пользователь не найден в базе", slog.Int("userID", userID))
			return false, nil
		}
		// Логируем ошибку выполнения запроса.
		slog.Error(op, "Ошибка при проверке пользователя в базе", slog.Int("userID", userID), slog.String("error", err.Error()))
		return false, err
	}

	// Логируем успешную проверку: найден ли пользователь в базе.
	slog.Info("Проверка пользователя в базе", slog.Int("userID", userID), slog.Bool("exists", count > 0))

	// Если count > 0, значит пользователь существует в базе.
	return count > 0, nil
}

// CreateUser - функция записи нового пользователя в базу данных.
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

// SaveUserKey - функция записи нового ключа в базу данных.
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

// IsKeyInDB - функция проверяет наличие ключа в базе данных по userID.
func (r *Repo) IsKeyInDB(userID int) (bool, error) {
	const op = "internal/bot/repository.go/IsKeyInDB"

	// Запрос на подсчёт строк с данным user_id в таблице keys.
	query := "SELECT COUNT(1) FROM keys WHERE user_id = $1"

	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		// Логируем ошибку выполнения запроса.
		slog.Error(op, "Ошибка при выполнении запроса", slog.Int("userID", userID), slog.String("error", err.Error()))
		return false, err
	}

	// Логируем успешную проверку: найден ли ключ в базе для этого user_id.
	slog.Info("Проверка наличия ключа в базе", slog.Int("userID", userID), slog.Bool("exists", count > 0))

	// Если count > 0, значит ключ существует в базе данных.
	return count > 0, nil
}

// GetExpirationTimeKey - функция получения времени истечения ключа пользователя.
func (r *Repo) GetExpirationTimeKey(userTgID int) (time.Time, error) {
	const op = "internal/bot/repository.go/GetExpirationTimeKey"

	var expiresAt time.Time
	query := `SELECT expires_at FROM public.keys WHERE user_id = $1`

	err := r.db.QueryRow(query, userTgID).Scan(&expiresAt)
	if err != nil {
		slog.Warn(op, slog.String("error", err.Error()))
		return expiresAt, err
	}

	return expiresAt, nil
}

// UpdateKeyExpiration - функция обновления(продления) времени истечения ключа пользователя.
func (r *Repo) UpdateKeyExpiration(userTgID int, newExpiration time.Time) error {
	const op = "internal/bot/repository.go/UpdateKeyExpiration"

	query := `UPDATE keys SET expires_at = $1 WHERE user_id = $2`

	_, err := r.db.Exec(query, newExpiration, userTgID)
	if err != nil {
		slog.Error(op, slog.String("error", err.Error()))
		return err
	}

	return nil
}

// GetConnectKeyOrReferral - функция для выдачи существующего ключа подключения.
func (r *Repo) GetConnectKeyOrReferral(userTgID int64, referral bool) (string, error) {
	const op = "internal/bot/repository.go/GetConnectionKey"

	// переменная для хранения запроса
	var query string

	// Формируем базовый запрос
	if referral {
		query = `SELECT referral_code FROM users WHERE telegram_id = $1`
	} else {
		query = `SELECT key FROM keys WHERE user_id = $1`
	}

	var connectionKeyOrReferral string
	err := r.db.QueryRow(query, userTgID).Scan(&connectionKeyOrReferral)

	// Проверяем, если ошибка - это отсутствие строк
	if err == sql.ErrNoRows {
		// Логируем, что ключ не найден
		slog.Info("Ключ подключения для пользователя не найден", slog.Int64("userID", userTgID))
		return "", nil
	}

	// Если ошибка другая, логируем её
	if err != nil {
		slog.Error(op, "Ошибка при получении ключа подключения", slog.Int64("userID", userTgID), slog.String("error", err.Error()))
		return "", err
	}

	// Возвращаем полученный ключ подключения
	return connectionKeyOrReferral, nil
}
