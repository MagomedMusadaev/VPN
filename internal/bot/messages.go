package bot

import (
	"bot_vpn/internal/entities"
	"bot_vpn/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"os"
	"strconv"
	"time"
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

	userID := update.Message.From.ID
	referralID := update.Message.CommandArguments() // Получаем реферальный ID

	var referredBy int
	if referralID != "" {
		refID, err := strconv.Atoi(referralID)
		if err != nil {
			slog.Error(op, "Ошибка преобразования referral_id", err)
			return
		}
		referredBy = refID
	}

	// запрос в базу
	exists, err := m.repo.IsUserInDB(int(userID))
	if err != nil {
		return
	}

	// Если пользователя нет в базе
	if !exists {
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

		if _, err = m.botAPI.Send(msg); err != nil {
			slog.Error(op, err)
		}

		// Создаём нового пользователя
		user := &entities.User{
			UserTgID:     int(userID),
			ChatTgID:     int(userID),
			ReferralCode: "https://t.me/keeper_vpn_bot?start=" + strconv.Itoa(int(userID)),
			CreatedAt:    time.Now(),
			ReferredBy:   referredBy,
		}

		// Сохраняем пользователя в базе данных
		if err = m.repo.CreateUser(user); err != nil {
			slog.Error(op, "Ошибка при создании пользователя в базе данных", err)
			return
		}

		return
	}

	var respMessage string
	expirationTime, err := m.repo.GetExpirationTimeKey(int(userID))

	// Проверяем наличие ошибки и истечение срока действия ключа.
	switch {
	case err != nil && errors.Is(err, sql.ErrNoRows):
		// У пользователя нет ключа
		respMessage = "У вас нет активного ключа. Вы можете приобрести подписку."

	case err == nil && time.Now().After(expirationTime):
		// Если ключ истёк
		respMessage = "Ваш ключ истёк. Продлите подписку, чтобы продолжить пользоваться сервисом."

	case err == nil:
		// Если всё в порядке
		respMessage = fmt.Sprintf("🔹 Ваша подписка активна до %s", expirationTime.Format("02.01.2006 15:04:05"))
	}

	welcomeMessage := fmt.Sprintf(
		"🎉 Добро пожаловать в Keeper VPN!\n\n"+
			"%s\n\n"+
			"✅ Оставайтесь под защитой без ограничений!",
		respMessage,
	)

	keyboard := m.keyBoard.GetStartButton()

	// Отправляем сообщение с клавиатурой
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, welcomeMessage)
	msg.ReplyMarkup = keyboard // Используем клавиатуру

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
		slog.Error(op, "Error sending message:", err)
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
			"Стоимость: %d рублей. 💳 \n",
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
func (m *MessengerBot) ManageUserDataAfterPayment(value, userTgID string) { // TODO переписать логику метода польностью
	const op = "internal/bot/messages.go/ManageUserKeyAfterPayment"

	// Преобразуем userTgID из строки в int
	intUserID, err := strconv.Atoi(userTgID)
	if err != nil {
		slog.Error(op, "Ошибка при преобразовании строки в int", err)
		return
	}

	// Устанавливаем время истечения в зависимости от стоимости
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
	}

	// Проверяем наличие ключа для пользователя в базе данных
	exists, err := m.repo.IsKeyInDB(intUserID)
	if err != nil {
		// Логируем ошибку и выходим из функции
		slog.Error("Ошибка при проверке наличия ключа", "userID", intUserID, "error", err)
		return
	}

	// Если ключа нет в базе
	if !exists {

		// Генерируем ключ для пользователя
		key, err := m.GenerateKey(userTgID)
		if err != nil && key == "" {
			return
		}

		// Формируем тело ключа для записи в базу
		keyRecord := &entities.Key{
			UserTgID:  intUserID,
			Key:       key + "#KeeperVPN",
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(expirationTime),
		}

		// Сохраняем ключ
		if err = m.repo.SaveUserKey(keyRecord); err != nil {
			return
		}

		slog.Info("Ключ успешно добавлен", "userID", intUserID)
		return
	}

	// Если ключ существует, обновляем срок действия
	expiresAt, err := m.repo.GetExpirationTimeKey(intUserID)
	if err != nil {
		return
	}

	// Определяем новое время истечения
	var newExpiration time.Time
	now := time.Now()

	if now.After(expiresAt) {
		// Если срок истёк, начинаем с текущего времени
		newExpiration = now.Add(expirationTime)
	} else {
		// Если подписка ещё активна, прибавляем к текущему сроку
		newExpiration = expiresAt.Add(expirationTime)
	}

	// Обновляем срок действия ключа
	if err = m.repo.UpdateKeyExpiration(intUserID, newExpiration); err != nil {
		return
	}

	// Формируем сообщение для пользователя о новом сроке действия
	message := fmt.Sprintf("Ваш ключ успешно продлён до %s.", newExpiration.Format("02.01.2006 15:04"))

	// Отправляем сообщение пользователю
	msg := tgbotapi.NewMessage(int64(intUserID), message)
	if _, err := m.botAPI.Send(msg); err != nil {
		// Логируем ошибку отправки сообщения
		slog.Error(op, "ошибка отправки сообщения:", err)
		return
	}

	slog.Info("Срок действия ключа успешно обновлён", slog.Int("user_id", intUserID))
}

// GenerateKey - функция генерации нового ключа
func (m *MessengerBot) GenerateKey(userTgID string) (string, error) {
	const op = "internal/bot/messages.go/GenerateKey"

	// Получаем URL API для запроса к серверу
	URL := os.Getenv("API_URL") // url сервера с outline
	if URL == "" {
		slog.Warn(op, "API_URL пуст")
		return "", errors.New("API_URL is empty") // Возвращаем ошибку, если переменная окружения пустая
	}

	payload := entities.NewKeyPayload(userTgID)
	slog.Info("Payload created",
		"userID:", userTgID,
		"time:", time.Now(),
	)

	// Формируем полный URL для запроса
	apiURL := fmt.Sprintf("%s/access-keys", URL)

	// Отправляем запрос на получение ключа
	key, err := m.httpRequest.SendKeyRequest(apiURL, payload)
	if err != nil {
		return "", err
	}

	// Формируем сообщение с ключом
	text := fmt.Sprintf("Ваш ключ доступа:\n```%s```", key+"#KeeperVPN")

	// Преобразуем userTgID в chatID (они одинаковы)
	chatID, err := strconv.ParseInt(userTgID, 10, 64)
	if err != nil {
		slog.Error(op, "Ошибка парсинга userTgID", err)
		return key, err // Возвращаем ключ, так как это не мешает его отправке
	}

	// Отправляем сообщение пользователю
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	if _, err = m.botAPI.Send(msg); err != nil {
		slog.Error(op, "Ошибка при отправке сообщения пользователю", err)
		return key, err // Возвращаем ключ, несмотря на ошибку отправки сообщения
	}

	// Логируем успешную отправку ключа
	slog.Info("Ключ успешно отправлен", "userID", userTgID)

	// Возвращаем полученный ключ
	return key, nil
}

// GetConnectStrOrReferral - функция для выдачи существующего ключа подключения или реферальной ссылки пользователю.
func (m *MessengerBot) GetConnectStrOrReferral(update tgbotapi.Update, referral bool) {
	const op = "internal/bot/messages.go/GetConnectStr"

	// Получаем ID пользователя из сообщения
	userTgID := update.Message.From.ID

	// Переменные для хранения ключа или ссылки
	var keyOrReferral string
	var err error

	// Проверяем, нужно ли возвращать реферальную ссылку
	if referral {
		// Получаем реферальную ссылку из репозитория
		keyOrReferral, err = m.repo.GetConnectKeyOrReferral(userTgID, true)
	} else {
		// Получаем ключ подключения из репозитория
		keyOrReferral, err = m.repo.GetConnectKeyOrReferral(userTgID, false)
	}

	// Если не найден, и ошибки нет
	if keyOrReferral == "" && err == nil {

		// Переменная для отправки сообщения с негативным исходом
		var text string

		if referral {
			// Логируем, что у пользователя нет рефералки
			slog.Info("У пользователя нет рефералки", slog.Int64("userID", userTgID))

			// Текст сообщения, что нет рефералки
			text = "Для получения реферальной ссылки, вам нужно хотя бы раз преобрести наш VPN"
		} else {
			// Логируем, что у пользователя нет ключа
			slog.Info("У пользователя нет ключа подключения", slog.Int64("userID", userTgID))

			// Текст сообщения, что ключ не найден
			text = "У вас нет активных ключей"
		}

		// Создаем и отправляем сообщение пользователю
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		if _, err := m.botAPI.Send(msg); err != nil {
			// Логируем ошибку отправки сообщения
			slog.Error(op, "ошибка отправки сообщения:", err)
		}
		return
	}

	// Если произошла ошибка при запросе
	if err != nil {
		// Логируем ошибку при получении ключа
		slog.Error(op, "Ошибка при получении ключа или рефералки", slog.Int64("userID", userTgID), slog.String("error", err.Error()))

		// Текст сообщения об ошибке
		text := "Произошла внутренняя ошибка, повторите запрос"

		// Создаем и отправляем сообщение пользователю
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
		if _, err := m.botAPI.Send(msg); err != nil {
			// Логируем ошибку отправки сообщения
			slog.Error(op, "ошибка отправки сообщения:", err)
		}
		return
	}

	// Переменная для отправки сообщения с успешным исходом
	var text string

	// Логируем успешную отправку ключа или реферальной ссылки
	if referral {
		text = fmt.Sprintf("Ваша реферальная ссылка:\n```%s```", keyOrReferral)
	} else {
		// Если ключ найден, отправляем его пользователю
		text = fmt.Sprintf("Ваш ключ доступа:\n```%s```", keyOrReferral)
	}

	// Создаем сообщение и устанавливаем ParseMode
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, text)
	msg.ParseMode = "Markdown" // Устанавливаем формат Markdown

	// Отправляем сообщение
	if _, err := m.botAPI.Send(msg); err != nil {
		// Логируем ошибку отправки сообщения
		slog.Error(op, "ошибка отправки сообщения:", err)
		return
	}

	// Логируем успешную отправку ключа или реферальной ссылки
	if referral {
		slog.Info("Реферальная ссылка успешно отправлена", slog.Int64("userID", userTgID))
	} else {
		slog.Info("Ключ успешно отправлен", slog.Int64("userID", userTgID))
	}
}

// Answer - функция для отправки сообщения продления ключа.
func (m *MessengerBot) Answer(callback *tgbotapi.CallbackQuery) {
	const op = "internal/bot/messages.go/Answer"

	// Если сообщение содержит текст, то удаляем его
	if callback.Message.Text != "" {
		editMsg := tgbotapi.NewEditMessageText(
			callback.Message.Chat.ID,
			callback.Message.MessageID,
			"", // Оставляем текст пустым, чтобы удалить старый
		)

		// Отправляем запрос на удаление текста
		if _, err := m.botAPI.Send(editMsg); err != nil {
			slog.Warn(op, err)
		}
	}

	// Получаем клавиатуру с кнопками
	startKeyboard := m.keyBoard.GetTariffKeyboard()

	// Обновляем клавиатуру с новой
	editMsgKeyboard := tgbotapi.NewEditMessageReplyMarkup(
		callback.Message.Chat.ID,
		callback.Message.MessageID,
		startKeyboard, // Новая клавиатура
	)

	// Отправляем новую клавиатуру
	if _, err := m.botAPI.Send(editMsgKeyboard); err != nil {
		slog.Error(op, "ошибка отправки клавиатуры:", err)
	}
}

//// TimeFunction - врEменная функция
//func (m *MessengerBot) TimeFunction(update tgbotapi.Update) {
//	messageText := fmt.Sprint("На стадии разработки!!!")
//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, messageText)
//	if _, err := m.botAPI.Send(msg); err != nil {
//		return
//	}
//}
