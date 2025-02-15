package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
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
	case "answer":
		h.messenger.Answer(callback)
	case "get_referral":
		h.messenger.GetConnectStrOrReferral(callback.From.ID, false)
	}
}

// Message - обрабатывает text-запросы
func (h *CallbackHandler) Message(update tgbotapi.Update) {

	switch {
	case update.Message != nil && strings.HasPrefix(update.Message.Text, "/start"):
		h.messenger.GetInfoStart(update)
	}

	switch update.Message.Text {
	//case "/start":
	//	h.messenger.GetInfoStart(update)
	case "/connect_str":
		h.messenger.GetConnectStrOrReferral(update.Message.From.ID, false)
	case "/instruction":
		h.messenger.GetInstruction(update)
	case "/referral":
		h.messenger.GetConnectStrOrReferral(update.Message.From.ID, true)
	case "/start referral":
		h.messenger.GetInfoStart(update)
	}
}
