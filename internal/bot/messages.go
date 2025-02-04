package bot

import (
	"bot_vpn/internal/entities"
	"bot_vpn/internal/utils"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
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
		"🛠️ Следуйте простым инструкциям\n\n"

	keyboard := m.keyBoard.GetTariffKeyboard()

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
	msg.ReplyMarkup = "Markdown" // Чтобы использовать жирный шрифт и эмодзи
	msg.ReplyMarkup = keyboard

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, err)
	}
}

// GetInstruction - отправляет пользователю инструкции по подключению к VPN через бота
func (m *MessengerBot) GetInstruction(update tgbotapi.Update) {
	const op = "internal/bot/messages.go/GetInstruction"

	// Текст инструкции с HTML-разметкой
	instructionText := "🌐 <b>Инструкция по подключению VPN</b>" +
		"<a href=\"https://telegra.ph/Nastrojka-VPN-cherez-Outline-01-30\">&#8203;</a>"

	// Создаем сообщение с инструкцией
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, instructionText)

	// Устанавливаем ParseMode для HTML
	msg.ParseMode = "HTML"

	// Отправляем сообщение
	_, err := m.botAPI.Send(msg)
	if err != nil {
		log.Println(op, "Error sending message:", err)
	}
}

// SendPaymentInfoWithButton - отправляет информацию о тарифе с кнопкой для перехода к оплате.
func (m *MessengerBot) SendPaymentInfoWithButton(callback *tgbotapi.CallbackQuery, month int) {
	const op = "internal/bot/messages.go/SendPaymentInfoWithButton"

	editMsg := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{}), // Убираем клавиатуру
	)

	if _, err := m.botAPI.Send(editMsg); err != nil {
		slog.Error(op, err)
	}

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
	userID := int(callback.From.ID)
	chatID := callback.Message.Chat.ID
	userName := callback.From.UserName
	reqMount := utils.GetMonthString(month)

	//Текст сообщения с информацией о тарифе
	text := fmt.Sprintf(
		"<b>💵 Оплата </b> \n"+
			"Вы выбрали тариф на %s. 💎 \n"+
			"Стоимость: %d рублей. 💳 \n"+
			"После оплаты ключ будет сгенерирован и отправлен автоматически. 🔑\n",
		reqMount, price,
	)

	// преобразуем в строки
	strPrice := strconv.Itoa(price)
	strUserID := strconv.Itoa(userID)

	// формируем url для оплаты
	paymentLink, err := m.CreatePayment(strPrice, strUserID, reqMount)
	if err != nil {
		return
	}

	// Кнопка для перехода на ссылку оплаты
	button := tgbotapi.NewInlineKeyboardButtonURL(fmt.Sprintf("Оплатить %d RUB 💳", price), paymentLink)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)

	// Создание и отправка сообщения с кнопкой
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, err)
		return
	}
	slog.Info("Данные для оплаты отправлены",
		"userID", userID,
		"userName", userName,
	)
}

// CreatePayment - формирует ссылку на оплату с metadata для ЮКассы
func (m *MessengerBot) CreatePayment(amount, tgUserID, reqMount string) (string, error) {
	const op = "internal/bot/messages.go/CreatePayment"

	// Получаем shopID и secretKey из ENV
	shopID := os.Getenv("YK_SHOP_ID")
	secretKey := os.Getenv("YK_SECRET_KEY")

	if shopID == "" || secretKey == "" {
		err := fmt.Errorf("shopID или secretKey не заданы в .env")
		slog.Error(op, err)
		return "", err
	}

	// Формируем тело запроса
	requestData := entities.PaymentRequest{
		Capture:     true,
		Description: "Оплата подписки на " + reqMount,
		Amount: entities.Amount{
			Value:    amount,
			Currency: "RUB",
		},
		Confirmation: entities.Confirmation{
			Type:      "redirect",
			ReturnURL: "https://zvuk.com/track/50970977",
		},
		Metadata: map[string]string{
			"user_tg_id": tgUserID,
		},
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		slog.Error("ошибка преобразования в json")
		return "", err
	}

	// Отправляем запрос
	paymentURL, err := m.httpRequest.GetPaymentURL(jsonData, shopID, secretKey, tgUserID)
	if err != nil {
		return "", err
	}

	return paymentURL, nil
}

// ManageUserDataAfterPayment - обрабатывает оплату пользователя и управляет его данными.
func (m *MessengerBot) ManageUserDataAfterPayment(value, userTgID string) {
	const op = "internal/bot/messages.go/ManageUserKeyAfterPayment"

	intUserID, err := strconv.Atoi(userTgID)
	if err != nil {
		slog.Error(op, "Ошибка при преобразовании строки в int", err)
		return
	}

	var expirationTime time.Duration
	switch value {
	case "100.00": // 1 месяц
		expirationTime = 30 * 24 * time.Hour
	case "190.00": // 2 месяца
		expirationTime = 60 * 24 * time.Hour
	case "270.00": // 3 месяца
		expirationTime = 90 * 24 * time.Hour
	case "490.00": // 6 месяцев
		expirationTime = 180 * 24 * time.Hour
	default:
		slog.Error(op, "Неизвестная сумма оплаты", "value", value)
		return
	}

	// Проверяем наличие пользователя в db
	exists, err := m.repo.IsUserInDB(intUserID)
	if err != nil {
		return
	}

	// Если пользователь не найден в db
	if !exists {
		// Создаём пользователя в таблице users
		user := &entities.User{
			UserTgID:     intUserID,
			ChatTgID:     intUserID,
			ReferralCode: "abs", // Если есть реферальный код, указываем его
			CreatedAt:    time.Now(),
			ReferredBy:   0, // Если пользователь был приглашён, указываем ID пригласившего
			//TODO: надо будет в db 0 id скипнуть
		}

		if err = m.repo.CreateUser(user); err != nil {
			slog.Error(op, "Ошибка при создании пользователя в базе данных", err)
			return
		}

		// Генерируем ключ и сохраняем его в таблице keys
		key, err := m.GenerateKey(userTgID)
		if err != nil && key == "" {
			return
		}

		keyRecord := &entities.Key{
			UserTgID:  intUserID,
			Key:       key,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(expirationTime),
		}

		if err = m.repo.SaveUserKey(keyRecord); err != nil {
			return
		}

		slog.Info("Пользователь и ключ успешно добавлены", "userID", intUserID, "key", key)
		return
	}

	// Если пользователь найден в db, то
	expiresAt, err := m.repo.GetExpirationTimeKey(intUserID)
	if err != nil {
		return
	}

	var newExpiration time.Time
	now := time.Now()

	if now.After(expiresAt) {
		// Если срок истёк, начинаем с текущего времени
		newExpiration = now.Add(expirationTime)
	} else {
		// Если подписка ещё активна, прибавляем к текущему сроку
		newExpiration = expiresAt.Add(expirationTime)
	}

	if err = m.repo.UpdateKeyExpiration(intUserID, newExpiration); err != nil {
		return
	}

	slog.Info("Срок действия ключа успешно обновлён", slog.Int("user_id", intUserID), slog.Time("new_expires_at", newExpiration))
}

// GenerateKey - функция генерации ключа
func (m *MessengerBot) GenerateKey(userTgID string) (string, error) {
	const op = "internal/bot/messages.go/GetKey"

	apiURL := os.Getenv("API_URL") // url сервера с outline
	if apiURL == "" {
		slog.Warn(op, "API_URL пуст")
		return "", errors.New("")
	}

	payload := entities.NewKeyPayload(userTgID)
	slog.Info("Payload created",
		"userID:", userTgID,
		"time:", time.Now(),
	)

	key, err := m.httpRequest.SendKeyRequest(apiURL, payload)
	if err != nil {
		return "", err
	}

	// Формируем сообщение с ключом
	text := fmt.Sprintf("Ваш ключ доступа:\n```%s```", key)

	chatID, err := strconv.ParseInt(userTgID, 10, 64)
	if err != nil {
		slog.Error(op, "ошибка парсинга userTgID", err)
		return key, err
	}

	// Отправляем сообщение
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err = m.botAPI.Send(msg); err != nil {
		slog.Error(op, "Failed to send message")
		return key, err
	}

	slog.Info("key sent successfully", userTgID)

	return key, nil
}

// TimeFunction - врEменная функция
func (m *MessengerBot) TimeFunction(update tgbotapi.Update) {
	messageText := fmt.Sprint("На стадии разработки!!!")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
	if _, err := m.botAPI.Send(msg); err != nil {
		return
	}
}
