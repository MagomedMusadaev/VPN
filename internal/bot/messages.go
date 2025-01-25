package bot

import (
	"bot_vpn/internal/entities"
	"bot_vpn/internal/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"os"
	"strconv"
	"time"
)

const (
	ttl = 4 * time.Hour
)

type Messenger interface {
	SendMessage(botAPI *tgbotapi.BotAPI, chatID int64, text string)
	SendMessageWithKeyboard(botAPI *tgbotapi.BotAPI, chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup)
	GetInfoStart(update tgbotapi.Update)
}

type MessengerBot struct {
	botAPI      *tgbotapi.BotAPI
	httpRequest *HttpRequest
	keyBoard    *KeyBoard
	repo        *Repo
}

func NewMessengerBot(botAPI *tgbotapi.BotAPI, httpRequest *HttpRequest, keyBoard *KeyBoard, repo *Repo) *MessengerBot {
	return &MessengerBot{
		botAPI:      botAPI,
		httpRequest: httpRequest,
		keyBoard:    keyBoard,
		repo:        repo,
	}
}

// GetInfoStart - обработчик для отправки приветственного сообщения и предоставления информации о тарифах.
func (m *MessengerBot) GetInfoStart(update tgbotapi.Update) {
	const op = "internal/bot/messages.go/GetInfoStart"

	welcomeMessage := "🎉 Приветствуем тебя в Keeper VPN! 🎉\n\n" +
		"🚀 Обеспечь себе безлимитный и быстрый VPN.\n\n" +
		"Для подключения:\n\n" +
		"📌 Выберите тариф\n" +
		"💳 Оплатите план\n" +
		"🛠️ Следуйте простым инструкциям\n\n" +
		"🟢 Вот ваши доступные пакеты:"

	keyboard := m.keyBoard.GetTariffKeyboard()

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
	msg.ReplyMarkup = "Markdown" // Чтобы использовать жирный шрифт и эмодзи
	msg.ReplyMarkup = keyboard

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, err)
	}
}

// SendPaymentInfoWithButton - отправляет информацию о тарифе с кнопкой для перехода к оплате.
func (m *MessengerBot) SendPaymentInfoWithButton(callback *tgbotapi.CallbackQuery, month int) {
	const op = "internal/bot/messages.go/SendPaymentInfoWithButton"

	// Массив с ценами для тарифов
	prices := []int{100, 190, 270, 490}

	// Расчёт конечной цены
	var price int
	if month == 6 {
		price = prices[3]
	} else {
		price = prices[month-1]
	}

	// Получение уникального идентификатора пользователя
	userID := callback.From.ID
	chatID := callback.Message.Chat.ID

	// Текст сообщения с информацией о тарифе
	text := fmt.Sprintf(
		"Вы выбрали тариф на %s. 💎\n"+
			"Стоимость: %d рублей. 💳\n\n"+
			"После оплаты ключ будет сгенерирован и отправлен автоматически. 🔑",
		utils.GetMonthString(month), price,
	)

	// Генерация ссылки на оплату с добавлением метаданных пользователя
	paymentLink := fmt.Sprintf(
		"https://example.com/payment?months=%d&price=%d&user_id=%d",
		month, price, userID, // TODO: с user_id разобраться
	)

	// Кнопка для перехода на ссылку оплаты
	button := tgbotapi.NewInlineKeyboardButtonURL("Оплатить 💳", paymentLink)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)

	// Создание и отправка сообщения с кнопкой
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, err)
		return
	}
	slog.Info("Данные для оплаты отправлены. Пользователь -", callback.From.ID, callback.From.UserName)

	// Проверяем наличие пользователя в БД
	exists, err := m.repo.IsUserInDB(userID)
	if err != nil {
		return
	}

	// Если пользователь не найден в БД, добавляем его временно в Redis
	if !exists {
		slog.Info("Пользователь не найден в базе данных, добавляем в Redis:", userID)

		err = m.repo.AddUserToRedis(fmt.Sprintf("%d", userID), fmt.Sprintf("%d", chatID), ttl) // TODO: user_id: struct{}
		if err != nil {
			return
		}
	}
}

func (m *MessengerBot) CheckAndUpdateUserKey(metadata *entities.MetaData) {
	const op = "internal/bot/messages.go/CheckAndUpdateUserKey"

	intUserID, err := strconv.ParseInt(metadata.Metadata.UserID, 10, 64)
	if err != nil {
		// Обработка ошибки, если преобразование не удалось
		slog.Error(op, "Ошибка при преобразовании строки в int64", slog.String("error", err.Error()))
		return
	}

	exists, err := m.repo.IsUserInDB(intUserID)
	if err == nil && exists {

		//m.repo.UpdateKeyExpiration(intUserID)
		// Если пользователь существует в базе данных, можем выполнить логику для обновления данных или продолжить выполнение программы.
		// В противном случае, если пользователя нет, добавляем нового пользователя со всеми данными.
	} else {
		// Если пользователя нет в базе данных или произошла ошибка при запросе, выполняем логику добавления нового пользователя.
		// Это может быть добавление нового пользователя в базу данных или сохранение его временно в Redis, как обсуждалось ранее.
	}
}

// TimeFunction - временная функция
func (m *MessengerBot) TimeFunction(update tgbotapi.Update) {
	messageText := fmt.Sprint("На стадии разработки")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	if _, err := m.botAPI.Send(msg); err != nil {
		return
	}
}

// нужно полностью переделать эту функцию (пока для тестинга)
func (m *MessengerBot) GetKey(update tgbotapi.Update) {
	const op = "internal/bot/messages.go/GetKey"

	apiURL := os.Getenv("API_URL") // url сервера с outline
	if apiURL == "" {
		slog.Warn(op, "API_URL пуст")
		return
	}

	userID := update.Message.From.ID
	strUserID := strconv.Itoa(int(userID))

	payload := entities.NewKeyPayload(strUserID)
	slog.Info("Payload created",
		"userID:", userID,
		"time:", time.Now(),
	)

	key, err := m.httpRequest.SendKeyRequest(apiURL, payload)
	if err != nil {
		return
	}

	messageText := fmt.Sprintf("Ваш ключ доступа: %s", key) // сделать чтобы можно было скопировать ключ

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, "Failed to send message")
		return
	}

	slog.Info("key sent successfully", userID)
}
