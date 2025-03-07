package entities

import "time"

// User - структура для хранения данны пользователя
type User struct {
	UserTgID     int       `db:"telegram_id"` // Telegram ID пользователя
	ChatTgID     int       `db:"chat_id"`     // Telegram ID пользователя
	ReferralCode string    `db:"referral_code"`
	CreatedAt    time.Time `db:"created_at"` // Дата первой записи данных пользователя
	ReferredBy   int       `db:"referred_by"`
	Nickname     string    `json:"nickname"`
}

// NewUser - конструктор для User
func NewUser(
	userTgID int,
	chatID int,
	referralCode string,
	createdAt time.Time,
	referredBy int,
	nickname string,
) User {
	return User{
		UserTgID:     userTgID,
		ChatTgID:     chatID,
		ReferralCode: referralCode,
		CreatedAt:    createdAt,
		ReferredBy:   referredBy,
		Nickname:     nickname,
	}
}
