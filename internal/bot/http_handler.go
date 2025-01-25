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
	var metadata entities.MetaData
	if err := json.Unmarshal(body, &metadata); err != nil {
		slog.Error(op, "Ошибка парсинга JSON:", slog.String("error", err.Error()))
		return
	}

	// Проверяем наличие обязательных данных
	if metadata.Metadata.UserID == "" || metadata.Metadata.Tariff == "" {
		slog.Error(op, "Отсутствуют обязательные поля метаданных")
		return
	}

	// Логика обработки платежа
	// 1. Проверяем наличие пользователя в БД

	//h.messenger.CheckAndUpdateUserKey(metadata)

	// 3. Если пользователя нет, получаем данные из Redis, создаём нового пользователя и возвращаем ключ.

	// TODO: Реализовать вызов сервиса (messages.go) для обработки платежа
}
