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
	slog.Info("Получена полезная нагрузка", slog.String("payload", string(body)))

	// Распаковываем JSON в структуру MetaData
	var event entities.Event
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error(op, "Ошибка парсинга JSON:", slog.String("error", err.Error()))
		return
	}

	switch event.Event {
	case "payment.succeeded":
		// Обрабатываем успешный платёж
		slog.Info("payment.succeeded:", event.Object.Metadata["user_tg_id"])
		if event.Object.Status == "succeeded" {
			// вызываем сервис слой
			h.messenger.ManageUserDataAfterPayment(event.Object.Amount.Value, event.Object.Metadata["user_tg_id"])
		}
	case "payment.canceled":
		// Обрабатываем отменённый платёж
		slog.Info("payment.canceled:", event.Object.Metadata["user_tg_id"])
		return
	default:
		slog.Info("Неизвестный ответ:", event.Object.Metadata["user_tg_id"])
		return
	}
}
