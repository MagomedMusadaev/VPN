package bot

import (
	"bot_vpn/internal/entities"
	"bot_vpn/internal/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"os"
	"strconv"
	"sync"
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
	client      *redis.Client
}

func NewMessengerBot(botAPI *tgbotapi.BotAPI, httpRequest *HttpRequest, keyBoard *KeyBoard, repo *Repo, client *redis.Client) *MessengerBot {
	return &MessengerBot{
		botAPI:      botAPI,
		httpRequest: httpRequest,
		keyBoard:    keyBoard,
		repo:        repo,
		client:      client,
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

// GetInstruction - отправляет пользователю инструкции по подключению к VPN через бота.
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

// CreatePayment - формирует ссылку на оплату с metadata для ЮКассы.
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
		Description: "Оплата на " + reqMount,
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
	const op = "internal/bot/messages.go/ManageUserDataAfterPayment"

	// Преобразуем userTgID из строки в int
	intUserID, err := strconv.Atoi(userTgID)
	if err != nil {
		slog.Error(op, "Ошибка при преобразовании строки в int", err)
		return
	}

	// Устанавливаем время истечения в зависимости от стоимости
	var referralExpirationTime time.Duration
	var expirationTime time.Duration

	switch value {
	case "100.00": // 1 месяц
		expirationTime = 30 * 24 * time.Hour
		referralExpirationTime = 4 * 24 * time.Hour
	case "190.00": // 2 месяца
		expirationTime = 60 * 24 * time.Hour
		referralExpirationTime = 8 * 24 * time.Hour
	case "270.00": // 3 месяца
		expirationTime = 90 * 24 * time.Hour
		referralExpirationTime = 12 * 24 * time.Hour
	case "490.00": // 6 месяцев
		expirationTime = 180 * 24 * time.Hour
		referralExpirationTime = 16 * 24 * time.Hour
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
		key, keyID, err := m.GenerateKey(userTgID)
		if err != nil && key == "" {
			return
		}

		// Формируем сообщение с ключом
		text := fmt.Sprintf("Ваш ключ доступа:\n```%s```", key+"#KeeperVPN")

		// Преобразуем userTgID в chatID (они одинаковы)
		chatID, err := strconv.ParseInt(userTgID, 10, 64)
		if err != nil {
			slog.Error(op, "Ошибка парсинга userTgID", err)
			return
		}

		// Отправляем сообщение пользователю
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		if _, err = m.botAPI.Send(msg); err != nil {
			slog.Error(op, "Ошибка при отправке сообщения пользователю", err)
			return
		}

		// Логируем успешную отправку ключа
		slog.Info("Ключ успешно отправлен", "userID", userTgID)

		// Формируем тело ключа для записи в базу
		keyRecord := &entities.Key{
			UserTgID:  intUserID,
			Key:       key + "#KeeperVPN",
			KeyID:     keyID,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(expirationTime),
		}

		// Сохраняем ключ
		if err = m.repo.SaveUserKey(keyRecord); err != nil {
			return
		}

		go m.AddReferralSubscriptionDays(userTgID, referralExpirationTime)

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

	// Получаем URL API для запроса к серверу
	url := os.Getenv("API_URL") // url сервера с outline
	if url == "" {
		slog.Warn(op, "API_URL пуст")
		return
	}

	// делаем запрос по user_id в таблицу keys и вытаскиваем keyID
	keyID, err := m.repo.GetKeyIDByUserID(userTgID)
	if err != nil {
		return
	}

	// Убираем ограничение по ключу в Outline Manager
	if err = m.httpRequest.RemoveOutlineKeyLimit(keyID, url); err != nil {
		return
	}

	if err = m.repo.UpdateProcessedKey(keyID, false); err != nil {
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

	// вызов функции которая будет добалять время рефералу 4 дня за месяц
	go m.AddReferralSubscriptionDays(userTgID, referralExpirationTime)
}

// AddReferralSubscriptionDays - функция, которая начисляет бонусные дни за реферала.
func (m *MessengerBot) AddReferralSubscriptionDays(userID string, expirationTime time.Duration) {
	const op = "internal/bot/messages.go/AddReferralSubscriptionDays"

	// Преобразуем userID в int
	intUserID, err := strconv.Atoi(userID)
	if err != nil {
		slog.Error(op, "Ошибка преобразования userID", err)
		return
	}

	// Получаем реферала пользователя
	referralUserID, err := m.repo.GetUserReferral(intUserID)
	if errors.Is(err, sql.ErrNoRows) {
		// Если реферала нет, просто выходим
		return
	}
	strReferralUserID := strconv.Itoa(referralUserID)

	// Проверяем наличие ключа у реферала
	exists, err := m.repo.IsKeyInDB(referralUserID)
	if err != nil {
		return
	}

	// Если ключ есть, обновляем срок его действия
	if exists {
		expiresAt, err := m.repo.GetExpirationTimeKey(referralUserID)
		if err != nil {
			return
		}

		var newExpiration time.Time
		now := time.Now()

		if now.After(expiresAt) {
			newExpiration = now.Add(expirationTime)
		} else {
			newExpiration = expiresAt.Add(expirationTime)
		}

		// Обновляем срок действия ключа в базе
		if err = m.repo.UpdateKeyExpiration(referralUserID, newExpiration); err != nil {
			return
		}

		// Уведомляем пользователя о продлении времени действия ключа
		if err = m.NotifyUserAboutReferralPurchase(strReferralUserID, newExpiration); err != nil {
			slog.Error("Ошибка при отправке уведомления о продлении ключа", "userID", referralUserID, "error", err)
			return
		}

		slog.Info("Время для реферала успешно обновлено")
		return
	}

	// Если ключа нет, создаём новый
	key, keyID, err := m.GenerateKey(strReferralUserID)
	if err != nil || key == "" {
		return
	}

	// Записываем новый ключ в базу
	keyRecord := &entities.Key{
		UserTgID:  referralUserID,
		Key:       key + "#KeeperVPN",
		KeyID:     keyID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(expirationTime),
	}

	if err = m.repo.SaveUserKey(keyRecord); err != nil {
		return
	}

	// Уведомляем пользователя о новом ключе
	if err = m.NotifyUserAboutReferralPurchase(strReferralUserID, time.Now().Add(expirationTime)); err != nil {
		return
	}

	slog.Info(op, "Ключ успешно добавлен и отправлен", slog.Int("userID", referralUserID))
}

// NotifyUserAboutReferralPurchase - отправляет уведомление пользователю о том, что по его реферальной ссылке был куплен тариф.
func (m *MessengerBot) NotifyUserAboutReferralPurchase(userTgID string, expiresAt time.Time) error {
	const op = "internal/bot/messenger.go/NotifyUserAboutReferralPurchase"

	// Преобразуем userTgID в chatID
	chatID, err := strconv.ParseInt(userTgID, 10, 64)
	if err != nil {
		slog.Error(op, slog.String("Ошибка парсинга userTgID", err.Error()))
		return err
	}

	// Формируем дату окончания подписки
	expirationDate := expiresAt.Format("02.01.2006 15:04")

	// Формируем сообщение
	text := fmt.Sprintf(
		"🎉 *По вашей реферальной ссылке оформлен тариф!* 🎉\n\n"+
			"📅 *Ключ действителен до:* `%s`\n\n"+
			"⬇️ Нажмите на кнопку ниже, чтобы получить ключ.",
		expirationDate,
	)

	// Создаём кнопку для активации команды
	button := tgbotapi.NewInlineKeyboardButtonData("🔑 Получить ключ", "get_referral")
	keyboard := tgbotapi.NewInlineKeyboardMarkup([]tgbotapi.InlineKeyboardButton{button})

	// Получаем ID последнего отправленного сообщения из Redis
	lastSentMessageID, err := m.repo.GetLastMessageID(chatID)
	if err != nil {
		return err
	}

	// Если есть старое сообщение, удаляем его
	if lastSentMessageID != 0 {
		deleteMsgConfig := tgbotapi.DeleteMessageConfig{
			ChatID:    chatID,
			MessageID: lastSentMessageID,
		}
		// Отправляем запрос на удаление старого сообщения
		_, err := m.botAPI.Request(deleteMsgConfig)
		if err != nil {
			slog.Error(op, slog.String("Ошибка при удалении старого сообщения", err.Error()))
			return err
		}
	}

	// Создаём и отправляем новое сообщение
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = keyboard

	sentMsg, err := m.botAPI.Send(msg)
	if err != nil {
		slog.Error(op, slog.String("Ошибка при отправке сообщения пользователю", err.Error()))
		return err
	}

	// Сохраняем ID отправленного сообщения в Redis
	err = m.repo.SaveLastMessageID(chatID, sentMsg.MessageID)
	if err != nil {
		return err
	}

	return nil
}

// GenerateKey - функция генерации нового ключа.
func (m *MessengerBot) GenerateKey(userTgID string) (string, string, error) {
	const op = "internal/bot/messages.go/GenerateKey"

	// Получаем URL API для запроса к серверу
	URL := os.Getenv("API_URL") // url сервера с outline
	if URL == "" {
		slog.Warn(op, "API_URL пуст")
		return "", "", errors.New("API_URL is empty") // Возвращаем ошибку, если переменная окружения пустая
	}

	payload := entities.NewKeyPayload(userTgID)
	slog.Info("Payload created",
		"userID:", userTgID,
		"time:", time.Now(),
	)

	// Формируем полный URL для запроса
	apiURL := fmt.Sprintf("%s/access-keys", URL)

	// Отправляем запрос на получение ключа
	key, keyID, err := m.httpRequest.SendKeyRequest(apiURL, payload)
	if err != nil {
		return "", "", err
	}

	// Возвращаем полученный ключ
	return key, keyID, nil
}

// GetConnectStrOrReferral - функция для выдачи существующего ключа подключения или реферальной ссылки пользователю.
func (m *MessengerBot) GetConnectStrOrReferral(userTgID int64, referral bool) {
	const op = "internal/bot/messages.go/GetConnectStr"

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
		msg := tgbotapi.NewMessage(userTgID, text)
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
		msg := tgbotapi.NewMessage(userTgID, text)
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
	msg := tgbotapi.NewMessage(userTgID, text)
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

// SetLimitForExpiredKeys - функция-ticket для установления лимита просроченных ключей.
func (m *MessengerBot) SetLimitForExpiredKeys() {
	const op = "internal/bot/messages.go/SetLimitForExpiredKeys"

	ticker := time.NewTicker(time.Hour * 12)
	defer ticker.Stop()

	// Получаем URL API для запроса к серверу
	url := os.Getenv("API_URL") // url сервера с outline
	if url == "" {
		slog.Warn(op, "API_URL пуст")
		return
	}

	for {
		select {
		case <-ticker.C:
			// Идем в базу и вытаскиваем key_id_outline для всех просроченных ключей
			expiredKeysID, err := m.repo.GetExpirationTimeKeysID()
			if err != nil {
				continue
			}

			// Параллельно делаем HTTP-запросы к Outline Manager и ставим лимиты
			var wg sync.WaitGroup
			for _, keyID := range expiredKeysID {
				wg.Add(1)
				go func(keyIDCopy int) {
					defer wg.Done()
					err = m.httpRequest.AddedOutlineKeyLimit(strconv.Itoa(keyIDCopy), url)
					if err != nil {
						return
					}
					if err = m.repo.UpdateProcessedKey(keyIDCopy, true); err != nil {
						return
					}
				}(keyID)
			}

			// Ждем завершения всех горутин
			wg.Wait()
		}
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
