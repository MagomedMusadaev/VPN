package entities

import "time"

// Key - структура для хранения информации о ключах доступа пользователя.
type Key struct {
	UserTgID  int       `db:"user_id"`    // ID пользователя в Telegram (связан с таблицей пользователей)
	Key       string    `db:"key"`        // Уникальный ключ доступа
	CreatedAt time.Time `db:"created_at"` // Дата и время создания ключа
	ExpiresAt time.Time `db:"expires_at"` // Дата и время истечения срока действия ключа
}
