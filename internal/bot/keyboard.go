package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// GetMainKeyboard возвращает основную клавиатуру с кнопками
func GetMainKeyboard() tgbotapi.InlineKeyboardMarkup {
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("Купить на 1 месяц", "buy_1m"),
			tgbotapi.NewInlineKeyboardButtonData("Купить на 6 месяцев", "buy_6m"),
		},
		{
			tgbotapi.NewInlineKeyboardButtonData("Купить на 1 год", "buy_1y"),
		},
	}
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}
