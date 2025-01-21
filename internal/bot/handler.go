package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandlerInt interface {
	Process(callback *tgbotapi.CallbackQuery)
	Message(update tgbotapi.Update)
}

// CallbackHandler обрабатывает callback-запросы
type CallbackHandler struct {
	messenger MessengerBot
}

// NewCallbackHandler создаёт новый экземпляр CallbackHandler
func NewCallbackHandler(messengerBot *MessengerBot) *CallbackHandler {
	return &CallbackHandler{
		messenger: *messengerBot,
	}
}

// Button обрабатывает callback-запросы
func (h *CallbackHandler) Button(callback *tgbotapi.CallbackQuery) {
	data := callback.Data

	switch data {
	case "buy_1m":
		h.messenger.SendMessage(callback.Message.Chat.ID, "Вы выбрали ключ на 1 месяц. Генерация...")
	case "buy_6m":
		h.messenger.SendMessage(callback.Message.Chat.ID, "Вы выбрали ключ на 6 месяцев. Генерация...")
	case "buy_1y":
		h.messenger.SendMessage(callback.Message.Chat.ID, "Вы выбрали ключ на 1 год. Генерация...")
	default:
		h.messenger.SendMessage(callback.Message.Chat.ID, "Неизвестная команда.")
	}
}

func (h *CallbackHandler) Message(update tgbotapi.Update) {
	//update.Message.Text
	switch {
	case update.Message.Text == "/start":
		h.messenger.GetRes(update)
	case update.Message.Text == "/getkey":
		h.messenger.GetKey(update)
	}
}
