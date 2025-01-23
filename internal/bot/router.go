package bot

import "net/http"

// InitRout - Функция инициализирует маршруты для HTTP-сервера.
func InitRout(handler *HttpHandler) {
	http.HandleFunc("/webhook", handler.PaymentWebhook) // TODO: изменить эндпоинт по мере необходимости.
}
