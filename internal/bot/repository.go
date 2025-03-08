package bot

import (
	"bot_vpn/internal/entities"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"strconv"
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

// GetUserReferral - функция выдаёт реферала пользователя по userID.
func (r *Repo) GetUserReferral(userID int) (int, error) {
	const op = "internal/bot/repository.go/GetUserReferral"

	query := `SELECT referred_by FROM users WHERE telegram_id = $1`
	var referralUserID int

	err := r.db.QueryRow(query, userID).Scan(&referralUserID)
	if err != nil {
		slog.Error(op, slog.String("Ошибка запроса к БД", err.Error()))
		return 0, err
	}

	if referralUserID == 0 {
		slog.Warn(op, slog.String("message", "Нет реферала у пользователя"), slog.Int("userID", userID))
		return 0, nil // Нет реферала — это не ошибка
	}

	return referralUserID, nil
}

// CreateUser - функция записи нового пользователя в базу данных.
func (r *Repo) CreateUser(user *entities.User) error {
	const op = "internal/bot/repository.go/CreateUser"

	query := `INSERT INTO users (telegram_id, chat_id, referral_code, created_at, referred_by, nickname) VALUES($1, $2, $3, $4, $5, $6)`

	_, err := r.db.Exec(query, user.UserTgID, user.ChatTgID, user.ReferralCode, user.CreatedAt, user.ReferredBy, user.Nickname)
	if err != nil {
		slog.Error(op, slog.String("error", err.Error()))
		return err
	}

	return nil
}

// SaveUserKey - функция записи нового ключа в базу данных.
func (r *Repo) SaveUserKey(key *entities.Key) error {
	const op = "internal/bot/repository.go/SaveUserKey"

	query := `INSERT INTO keys (user_id, key, created_at, expires_at, key_id_outline) VALUES($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(query, key.UserTgID, key.Key, key.CreatedAt, key.ExpiresAt, key.KeyID)
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

// GetExpirationTimeKeysID - функция для получения keyID ключей, которые протухли.
func (r *Repo) GetExpirationTimeKeysID() ([]int, error) {
	const op = "internal/bot/repository.go/GetExpirationTimeKeysID"

	query := `SELECT key_id_outline FROM keys WHERE expires_at < NOW() AND processed = false`

	rows, _ := r.db.Query(query)
	defer rows.Close()

	// Проверяем, есть ли данные в rows
	if !rows.Next() {
		slog.Info("Нет протухших ключей")
		return []int{}, nil
	}

	// Если данные есть, начинаем обработку
	var expiredKeys []int
	var keyID int

	// Так как мы уже один раз вызвали rows.Next(), сначала обрабатываем первую строку
	if err := rows.Scan(&keyID); err != nil {
		slog.Error(op, "Ошибка чтения key_id_outline", err)
		return nil, err
	}
	expiredKeys = append(expiredKeys, keyID)

	// Далее продолжаем читать остальные строки
	for rows.Next() {
		if err := rows.Scan(&keyID); err != nil {
			slog.Error(op, "Ошибка чтения key_id_outline", err)
			return nil, err
		}
		expiredKeys = append(expiredKeys, keyID)
	}

	if err := rows.Err(); err != nil {
		slog.Error(op, "Ошибка при итерации по строкам", err)
		return nil, err
	}

	return expiredKeys, nil
}

// CheckExpiringKeysAndNotify - функция для получения userID юзеров у которых скоро протухнет ключ
func (r *Repo) CheckExpiringKeysAndNotify() ([]int, error) {
	const op = "internal/bot/repository.go/CheckExpiringKeysAndNotify"

	query := `SELECT user_id FROM keys WHERE expires_at <= NOW() + INTERVAL '24 HOURS'
			AND expiration_notified = FALSE`

	rows, err := r.db.Query(query)
	if err != nil {
		err = fmt.Errorf("oшибка запроса в базу: %w", err)
		slog.Error(op, err.Error())
		return nil, err
	}
	defer rows.Close()

	// Слайс для хранения user_id с истекающими ключами
	var userIDs []int

	// Чтение результатов запроса
	for rows.Next() {
		var userID int
		if err = rows.Scan(&userID); err != nil {
			slog.Error(op, "Ошибка сканирования ответа", slog.String("error", err.Error()))
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}

	// Проверка на наличие ошибок при итерации по строкам
	if err = rows.Err(); err != nil {
		slog.Error(op, "Ошибка при проверке сканирования данных", slog.String("error", err.Error()))
		return nil, err
	}
	return userIDs, nil
}

// UpdateExpiringKey - функция которая меняет флажок expiration_at для ключа.
func (r *Repo) UpdateExpiringKey(userID int, flag bool) error {
	const op = "internal/bot/repository.go/UpdateExpiringKey"

	query := `UPDATE keys SET expiration_notified = $1 WHERE user_id = $2`

	result, err := r.db.Exec(query, flag, userID)
	if err != nil {
		slog.Error(op, "Ошибка при обновлении expiration_at", err)
		return err
	}

	// Проверяем, были ли обновлены строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(op, "Ошибка получения количества обновленных строк", err)
		return err
	}

	if rowsAffected == 0 {
		slog.Warn(op, "Не найден ключ с таким userID", "userID", userID)
		return fmt.Errorf("ключ с userID=%d не найден", userID)
	}

	slog.Info("Флажок expiration_notified успешно обновлён",
		slog.Int("userID", userID),
		slog.Bool("flag", flag),
	)
	return nil
}

// UpdateProcessedKey - функция которая меняет флажок processed для ключа.
func (r *Repo) UpdateProcessedKey(keyID int, flag bool) error {
	const op = "internal/bot/repository.go/UpdateProcessedKey"

	query := `UPDATE keys SET processed = $1 WHERE key_id_outline = $2`

	result, err := r.db.Exec(query, flag, keyID)
	if err != nil {
		slog.Error(op, "Ошибка при обновлении processed", err)
		return err
	}

	// Проверяем, были ли обновлены строки
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(op, "Ошибка получения количества обновленных строк", err)
		return err
	}

	if rowsAffected == 0 {
		slog.Warn(op, "Не найден ключ с таким key_id_outline", "keyID", keyID)
		return fmt.Errorf("ключ с key_id_outline=%d не найден", keyID)
	}

	slog.Info("Флажок processed успешно обновлён",
		slog.Int("keyID", keyID),
		slog.Bool("flag", flag),
	)
	return nil
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

// GetKeyIDByUserID - функция для получения id ключа в outline.
func (r *Repo) GetKeyIDByUserID(userID string) (int, error) {
	const op = "internal/bot/repository.go/GetKeyIDByUserID"

	query := `SELECT key_id_outline FROM keys WHERE user_id = $1`

	var keyID int
	if err := r.db.QueryRow(query, userID).Scan(&keyID); err != nil {
		slog.Error(op, "Ошибка выдачи keyID по userID", err)
		return 0, err
	}

	return keyID, nil
}

// SaveLastMessageID - сохраняет ID последнего отправленного сообщения в Redis
func (r *Repo) SaveLastMessageID(chatID int64, messageID int) error {
	const op = "internal/bot/repository.go/SaveLastMessageID"

	// Удаляем старое сообщение, если оно есть
	if err := r.client.Del(context.Background(), fmt.Sprintf("last_message:%d", chatID)).Err(); err != nil {
		slog.Error(op, "Ошибка при удалении старого сообщения", err)
		return err
	}

	// Сохраняем новое сообщение в Redis с TTL 12 часов
	err := r.client.Set(context.Background(), fmt.Sprintf("last_message:%d", chatID), messageID, 12*time.Hour).Err()
	if err != nil {
		slog.Error(op, "Ошибка при сохранении нового сообщения", err)
		return err
	}

	return nil
}

// GetLastMessageID - возвращает ID последнего сообщения для пользователя из Redis
func (r *Repo) GetLastMessageID(chatID int64) (int, error) {
	const op = "internal/bot/repository.go/GetLastMessageID"

	messageIDStr, err := r.client.Get(context.Background(), fmt.Sprintf("last_message:%d", chatID)).Result()
	if errors.Is(err, redis.Nil) {
		slog.Warn(op, slog.String("Нет данных по ключу в redis", ""))
		return 0, nil // Сообщение не найдено
	}
	if err != nil {
		slog.Error(op, "Ошибка при получении ID последнего сообщения", err)
		return 0, err
	}

	// Преобразуем строку в int
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		slog.Error(op, "Ошибка при преобразовании messageID в int", err)
		return 0, err
	}

	return messageID, nil
}
