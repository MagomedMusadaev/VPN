package bot

import (
	"bot_vpn/internal/entities"
	"bot_vpn/internal/utils"
	"encoding/json"
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
	userID := callback.From.ID
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
	strUserID := strconv.Itoa(int(userID))

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

	// Проверяем наличие пользователя в БД
	exists, err := m.repo.IsUserInDB(userID)
	if err != nil {
		return
	}

	// Если пользователь не найден в БД, добавляем его временно в Redis
	if !exists {
		slog.Info("Пользователь не найден в db, добавляем в Redis:",
			"userID", userID,
			"userName", userName,
		)

		err = m.repo.AddUserToRedis(
			fmt.Sprintf("%d", userID),
			fmt.Sprintf("%d", chatID),
			ttl,
		)
		if err != nil {
			return
		}
	}
}

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
			ReturnURL: "https://meet.google.com/sju-ktir-vze",
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

// ManageUserKeyAfterPayment обрабатывает оплату пользователя и управляет его ключом доступа.
//func (m *MessengerBot) ManageUserKeyAfterPayment(metadata entities.MetaData) {
//	const op = "internal/bot/messages.go/ManageUserKeyAfterPayment"
//
//	intUserID, err := strconv.ParseInt(metadata.Metadata.UserID, 10, 64)
//	if err != nil {
//		slog.Error(op, "Ошибка при преобразовании строки в int64", slog.String("error", err.Error()))
//		return
//	}
//
//	exists, err := m.repo.IsUserInDB(intUserID)
//	if err == nil && exists {
//
//		//m.repo.UpdateKeyExpiration(intUserID)
//		// Если пользователь существует в базе данных, можем выполнить логику для обновления данных или продолжить выполнение программы.
//		// В противном случае, если пользователя нет, добавляем нового пользователя со всеми данными.
//	} else {
//		// Если пользователя нет в базе данных или произошла ошибка при запросе, выполняем логику добавления нового пользователя.
//		// Это может быть добавление нового пользователя в базу данных или сохранение его временно в Redis, как обсуждалось ранее.
//	}
//	// Функция выполняет следующие шаги:
//	// 1. Проверяет, существует ли пользователь в базе данных:
//	//    - Если пользователь существует, обновляет время истечения его ключа доступа в соответствии с продлённой подпиской.
//	//    - Если пользователь не найден, создаёт новую запись в базе данных с необходимыми данными.
//	// 2. Генерирует или обновляет ключ доступа для пользователя.
//	// 3. Отправляет обновлённый или вновь созданный ключ доступа пользователю.
//
//}

// TimeFunction - врEменная функция
func (m *MessengerBot) TimeFunction(update tgbotapi.Update) {
	messageText := fmt.Sprint("На стадии разработки!!!")
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

	// Формируем сообщение с ключом
	text := fmt.Sprintf("Ваш ключ доступа:\n```%s```", key)

	// Отправляем сообщение
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown"

	if _, err := m.botAPI.Send(msg); err != nil {
		slog.Error(op, "Failed to send message")
		return
	}

	slog.Info("key sent successfully", userID)
}
