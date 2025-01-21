package bot

import (
	"bot_vpn/internal/entities"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	"os"
)

type Messenger interface {
	SendMessage(botAPI *tgbotapi.BotAPI, chatID int64, text string)
	SendMessageWithKeyboard(botAPI *tgbotapi.BotAPI, chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup)
}

type MessengerBot struct {
	botAPI      *tgbotapi.BotAPI
	httpHandler *HttpHandler
}

func NewMessengerBot(botAPI *tgbotapi.BotAPI, httpHandler *HttpHandler) *MessengerBot {
	return &MessengerBot{
		botAPI:      botAPI,
		httpHandler: httpHandler,
	}
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

func (m *MessengerBot) GetKey(update tgbotapi.Update) {
	const op = "internal/bot/messages.go/GetKey"

	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		slog.Warn(op, "API_URL пуст")
		return
	}

	userID := update.Message.From.ID
	userName := update.Message.From.UserName
	firstName := update.Message.From.FirstName

	payload := entities.NewKeyPayload(userID, userName, firstName)
	slog.Info("Payload created",
		"userID", userID,
		"userName", userName,
		"firstName", firstName,
	)

	key, err := m.httpHandler.SendKeyRequest(apiURL, payload)
	if err != nil {
		return
	}

	messageText := fmt.Sprintf("Ваш ключ доступа: %s", key)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, "Failed to send message")
		return
	}

	slog.Info(op, "key sent successfully")
}
