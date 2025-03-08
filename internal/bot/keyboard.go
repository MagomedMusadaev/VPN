package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type KeyBoardInt interface {
	GetTariffKeyboard() tgbotapi.InlineKeyboardMarkup
	GetStartButtonWithKey() tgbotapi.InlineKeyboardMarkup
	GetStartButtonWithoutKey() tgbotapi.InlineKeyboardMarkup
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

// GetStartButtonWithKey -  возвращает клавиатуру с кнопками если user уже является пользователем.
func (k *KeyBoard) GetStartButtonWithKey() tgbotapi.InlineKeyboardMarkup {
	// Создаём кнопки.
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 Продлить", "answer"),
		},
	}

	// Возвращаем разметку с кнопками.
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

// GetStartButtonWithoutKey -  возвращает клавиатуру с кнопками если user ещё не является пользователем.
func (k *KeyBoard) GetStartButtonWithoutKey() tgbotapi.InlineKeyboardMarkup {
	// Создаём кнопки.
	buttons := [][]tgbotapi.InlineKeyboardButton{
		{
			tgbotapi.NewInlineKeyboardButtonData("💰 Купить", "answer"),
		},
	}

	// Возвращаем разметку с кнопками.
	return tgbotapi.NewInlineKeyboardMarkup(buttons...)
}

func (k *KeyBoard) GetHelpButton() tgbotapi.InlineKeyboardMarkup {
	// Создаём inline-кнопку с ссылкой на поддержку
	supportButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("🔗 Нажмите здесь, чтобы получить помощь!",
				"https://t.me/Keeper_vpn_support"),
		),
	)

	return supportButton
}
