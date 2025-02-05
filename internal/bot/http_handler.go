package bot

import (
	"bot_vpn/internal/entities"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type HttpHandlerInt interface {
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
		slog.Error(op, "Неверный метод запроса", slog.String("method", r.Method))
		http.Error(w, "Неверный метод запроса", http.StatusMethodNotAllowed)
		return
	}

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error(op, "Ошибка при чтении тела запроса", slog.String("error", err.Error()))
		http.Error(w, "Ошибка при обработке запроса", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Логируем тело запроса для отладки (можно ограничить на проде, если нужно)
	slog.Info("Получена полезная нагрузка", slog.String("payload", string(body)))

	// Распаковываем JSON в структуру MetaData
	var event entities.Event
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error(op, "Ошибка парсинга JSON", slog.String("error", err.Error()))
		http.Error(w, "Ошибка при обработке данных", http.StatusBadRequest)
		return
	}

	// Обрабатываем события
	switch event.Event {
	case "payment.succeeded":
		// Обрабатываем успешный платёж
		userTgID := event.Object.Metadata["user_tg_id"]
		slog.Info("Платёж успешен", slog.String("user_tg_id", userTgID))

		if event.Object.Status == "succeeded" {
			// вызываем сервисный слой для обработки данных
			h.messenger.ManageUserDataAfterPayment(event.Object.Amount.Value, userTgID)
		}
	case "payment.canceled":
		// Обрабатываем отменённый платёж
		userTgID := event.Object.Metadata["user_tg_id"]
		slog.Info("Платёж отменён", slog.String("user_tg_id", userTgID))
		return
	default:
		// Неизвестный тип события
		userTgID := event.Object.Metadata["user_tg_id"]
		slog.Warn("Неизвестный ответ от платежной системы", slog.String("user_tg_id", userTgID), slog.String("event", event.Event))
		http.Error(w, "Неизвестное событие", http.StatusBadRequest)
		return
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
}
