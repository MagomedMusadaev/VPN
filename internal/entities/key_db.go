package entities

import "time"

// Key - структура для хранения информации о ключах доступа пользователя.
type Key struct {
	UserTgID  int       `db:"telegram_id"` // Используем telegram_id в качестве внешнего ключа
	Key       string    `db:"key"`
	CreatedAt time.Time `db:"created_at"`
	ExpiresAt time.Time `db:"expires_at"`
}
