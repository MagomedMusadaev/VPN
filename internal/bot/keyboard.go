package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type KeyBoardInt interface {
	GetTariffKeyboard() tgbotapi.InlineKeyboardMarkup
}

type KeyBoard struct {
}

func NewKeyBoard() *KeyBoard {
	return &KeyBoard{}
}

// GetTariffKeyboard -  возвращает основную клавиатуру с кнопками.
func (k *KeyBoard) GetTariffKeyboard() tgbotapi.InlineKeyboardMarkup {
	// Создаём кнопки для тарифов.
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 100 ₽ за 1 месяц", "buy_1m"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 190 ₽ за 2 месяца", "buy_2m"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 270 ₽ за 3 месяца", "buy_3m"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 490 ₽ за 6 месяцев", "buy_6m"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("🔗 Реферальная ссылка", "get_referral"),
		},
	}

	// Возвращаем разметку с кнопками.
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// GetStartButton -  возвращает клавиатуру с кнопками если user уже является пользователем.
func (k *KeyBoard) GetStartButton() tgbotapi.InlineKeyboardMarkup {
	// Создаём кнопки.
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 Продлить", "answer"),
		},
	}

	// Возвращаем разметку с кнопками.
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
