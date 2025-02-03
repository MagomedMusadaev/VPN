package bot

import (
	"bot_vpn/internal/entities"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type HttpHandlerInt interface {
	SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, error)
	PaymentWebhook(w http.ResponseWriter, r *http.Request)
}

type HttpHandler struct {
	messenger *MessengerBot
}

func NewHttpHandler(messenger *MessengerBot) *HttpHandler {
	return &HttpHandler{
		messenger: messenger,
	}
}

// PaymentWebhook - обрабатывает вебхук платежной системы.
func (h *HttpHandler) PaymentWebhook(w http.ResponseWriter, r *http.Request) {
	const op = "internal/bot/http_handler/PaymentWebhook"

	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		slog.Error(op, "Неверный метод запроса")
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(op, "Ошибка при чтении тела запроса:", err)
		return
	}

	// Логируем тело запроса для отладки
	slog.Info(op, "Получена полезная нагрузка", slog.String("payload", string(body)))

	// Распаковываем JSON в структуру MetaData
	var event entities.Event
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error(op, "Ошибка парсинга JSON:", slog.String("error", err.Error()))
		return
	}

	switch event.Status {
	case "payment.succeeded":
		// Обрабатываем успешный платёж
		slog.Info("payment.succeeded:", event)
		if userTgID, ok := event.Metadata["user_tg_id"]; ok {
			slog.Info("user_tg_id:", userTgID)
		} else {
			slog.Info("user_tg_id не найден")
		}
	case "payment.canceled":
		// Обрабатываем отменённый платёж
		slog.Info("payment.canceled:", event)
		if userTgID, ok := event.Metadata["user_tg_id"]; ok {
			slog.Info("user_tg_id:", userTgID)
		} else {
			slog.Info("user_tg_id не найден")
		}
	default:
		slog.Info("Неизвустный ответ:", event)
	}

	//switch event.Status {
	//case "payment.succeeded":
	//	// Обрабатываем успешный платёж
	//	//processSuccessfulPayment(event.Object)
	//
	//	slog.Info("payment.succeeded:", event)
	//	slog.Info("payment.succeeded:", event.Metadata["user_tg_id"])
	//
	//case "payment.canceled":
	//	// Обрабатываем отменённый платёж
	//	//processCancelledPayment(event.Object)
	//
	//	slog.Info("payment.succeeded:", event)
	//	slog.Info("payment.succeeded:", event.Metadata["user_tg_id"])
	//
	//// Добавьте другие обработчики для других типов событий
	//default:
	//	http.Error(w, "Неизвестный тип события", http.StatusBadRequest)
	//}

	// Логика обработки платежа
	// 1. Проверяем наличие пользователя в БД

	//h.messenger.ManageUserKeyAfterPayment(metadata)

	// 2. Если пользователя нет, получаем данные из Redis, создаём нового пользователя и возвращаем ключ.
}
