package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CallbackHandlerInt interface {
	Process(callback *tgbotapi.CallbackQuery)
	Message(update tgbotapi.Update)
}

type CallbackHandler struct {
	messenger *MessengerBot
}

func NewCallbackHandler(messengerBot *MessengerBot) *CallbackHandler {
	return &CallbackHandler{
		messenger: messengerBot,
	}
}

// Button - обрабатывает callback-запросы
func (h *CallbackHandler) Button(callback *tgbotapi.CallbackQuery) {
	data := callback.Data

	switch data {
	case "buy_1m":
		h.messenger.SendPaymentInfoWithButton(callback, 1)
	case "buy_2m":
		h.messenger.SendPaymentInfoWithButton(callback, 2)
	case "buy_3m":
		h.messenger.SendPaymentInfoWithButton(callback, 3)
	case "buy_6m":
		h.messenger.SendPaymentInfoWithButton(callback, 6)
	case "get_referral":
		//h.messenger.SendMessage(callback.Message.Chat.ID, "реферальная ссылка")
	default:
		//h.messenger.SendMessage(callback.Message.Chat.ID, "Ты дурак что ли!?")
	}
}

// Message - обрабатывает text-запросы
func (h *CallbackHandler) Message(update tgbotapi.Update) {

	switch update.Message.Text {
	case "/start":
		h.messenger.GetInfoStart(update)
	case "/daykey":
		h.messenger.GetKey(update)
	case "/connect_str":
		h.messenger.TimeFunction(update)
	case "/instruction":
		h.messenger.GetInstruction(update)
	case "/referral":
		h.messenger.TimeFunction(update)
	}
}
