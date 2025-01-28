package entities

import "time"

// User представляет пользователя в системе Telegram и связанную с ним информацию о ключах.
type User struct {
	UserTgID      int64     `json:"user_tg_id" db:"user_tg_id"`           // Telegram ID пользователя
	KeyID         int       `json:"key_id" db:"key_id"`                   // ID ключа в системе Outline Manager
	Key           string    `json:"key" db:"key"`                         // Ключ подключения
	KeyStatus     string    `json:"key_status" db:"key_status"`           // Статус ключа (например, активен, истёк, ограничен)
	KeyStartDate  time.Time `json:"key_start_date" db:"key_start_date"`   // Дата начала действия ключа
	KeyExpiryDate time.Time `json:"key_expiry_date" db:"key_expiry_date"` // Дата истечения действия ключа
	CreatedAt     time.Time `json:"created_at" db:"created_at"`           // Дата первой записи данных пользователя
}

// NewUser - конструктор для User
func NewUser(
	userTgID int64,
	keyID int, key,
	keyStatus string,
	KeyStartDate,
	KeyExpiryDate,
	CreatedAt time.Time,
) User {
	return User{
		UserTgID:      userTgID,
		KeyID:         keyID,
		Key:           key,
		KeyStatus:     keyStatus,
		KeyStartDate:  KeyExpiryDate,
		KeyExpiryDate: KeyStartDate,
		CreatedAt:     CreatedAt,
	}
}
