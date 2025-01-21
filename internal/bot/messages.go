package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

type Messenger interface {
	SendMessage(botAPI *tgbotapi.BotAPI, chatID int64, text string)
	SendMessageWithKeyboard(botAPI *tgbotapi.BotAPI, chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup)
}

type MessengerBot struct {
	botAPI *tgbotapi.BotAPI
}

func NewMessengerBot(botAPI *tgbotapi.BotAPI) *MessengerBot {
	return &MessengerBot{botAPI: botAPI}
}

// SendMessage отправляет текстовое сообщение
func (m *MessengerBot) SendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := m.botAPI.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

// SendMessageWithKeyboard отправляет сообщение с клавиатурой
func (m *MessengerBot) SendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	_, err := m.botAPI.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения с клавиатурой: %v", err)
	}
}

func (m *MessengerBot) GetRes(update tgbotapi.Update) {
	// Создаем сообщение
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Хотя бы МЯУ скажите!")

	// Отправляем сообщение с использованием метода SendMessage
	m.SendMessage(update.Message.Chat.ID, msg.Text)
}
