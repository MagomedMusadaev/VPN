package bot

import (
	"bot_vpn/internal/entities"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type HttpRequestInt interface {
	SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, error)
	GetPaymentURL(jsonData []byte, shopID, secretKey, tgUserID string) (string, error)
	RemoveOutlineKeyLimit(userID int, url string) error
}

type HttpRequest struct {
}

func NewHttpRequest() *HttpRequest {
	return &HttpRequest{}
}

// createInsecureHTTPClient - создает и возвращает HTTP-клиент с отключенной проверкой SSL. TODO: избавиться от этого в концее
func createInsecureHTTPClient() *http.Client {
	// Настраиваем транспорт с отключенной проверкой сертификатов SSL
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Отключаем проверку SSL
		},
	}

	// Создаем HTTP-клиент с указанным транспортом и тайм-аутом
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Minute, // Устанавливаем тайм-аут
	}
	return client
}

// SendKeyRequest - функция генерации ключа подключения.
func (h *HttpRequest) SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, string, error) { // TODO: нужно дорабоать (возможно)
	const op = "internal/bot/http_request/SendKeyRequest"

	// Сериализация полезной нагрузки в JSON.
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error(op, "Ошибка сериализации полезной нагрузки", slog.String("error", err.Error()))
		return "", "", err
	}

	// Создание нового HTTP-запроса.
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error(op, "Ошибка создания HTTP-запроса", slog.String("error", err.Error()))
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Используем HTTP-клиент с отключенной проверкой SSL.
	client := createInsecureHTTPClient()

	// Выполнение HTTP-запроса.
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(op, "Ошибка выполнения HTTP-запроса", slog.String("error", err.Error()))
		return "", "", err
	}
	defer resp.Body.Close()

	// Проверка статуса ответа.
	if resp.StatusCode != http.StatusCreated {
		slog.Error(op, "Неверный статус ответа", slog.Int("status_code", resp.StatusCode))
		return "", "", fmt.Errorf("получен статус %d, ожидался %d", resp.StatusCode, http.StatusCreated)
	}

	// Распаковка ответа.
	var respBody entities.RespBodyKey
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.Error(op, "Ошибка декодирования тела ответа", slog.String("error", err.Error()))
		return "", "", err
	}

	// Возвращаем ключ из ответа.
	return respBody.Key, respBody.ID, nil
}

// AddedOutlineKeyLimit - функция для установления ограничения ключа подключения.
func (h *HttpRequest) AddedOutlineKeyLimit(keyID string, url string) error {
	const op = "internal/bot/http_request/AddedOutlineKeyLimit"

	// Формируем полный URL для запроса
	apiURL := fmt.Sprintf("%s/access-keys/%s/data-limit", url, keyID)

	fmt.Println("ВОТ:", apiURL)

	// Формируем тело запроса с ограничением (например, 10000 байт).
	requestBody := `{"limit": {"bytes": 1048576}}` // 1 МБ

	// Создание нового HTTP-запроса.
	req, err := http.NewRequest("PUT", apiURL, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		slog.Error(op, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Используем HTTP-клиент с отключенной проверкой SSL.
	client := createInsecureHTTPClient()

	// Выполнение HTTP-запроса.
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(op, err)
		return err
	}
	defer resp.Body.Close()

	// Проверка статуса ответа.
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New(fmt.Sprintf("неожиданный статус: %d", resp.StatusCode))
		slog.Error(op, err)
		return err
	}

	slog.Info("Ограничение по ключу %d успешно установлено\n", keyID)
	return nil
}

// RemoveOutlineKeyLimit - функция для отмены ограничения ключа подключения.
func (h *HttpRequest) RemoveOutlineKeyLimit(keyID int, url string) error {
	const op = "internal/bot/http_request/RemoveOutlineKeyLimit"

	// Формируем полный URL для запроса
	apiURL := fmt.Sprintf("%s/access-keys/%d/data-limit", url, keyID)

	// Создание нового HTTP-запроса.
	req, err := http.NewRequest("DELETE", apiURL, nil)
	if err != nil {
		slog.Error(op, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Используем HTTP-клиент с отключенной проверкой SSL.
	client := createInsecureHTTPClient()

	// Выполнение HTTP-запроса.
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(op, err)
		return err
	}
	defer resp.Body.Close()

	// Проверка статуса ответа.
	if resp.StatusCode == http.StatusNotFound {
		err = errors.New(fmt.Sprintf("неожиданный статус: %d", resp.StatusCode))
		slog.Error(op, err)
		return err
	}

	slog.Info("Ограничение по ключу %d успешно снято\n", keyID)
	return nil
}

// GetPaymentURL - функци генерации ссылки оплаты на ЮКасса.
func (h *HttpRequest) GetPaymentURL(jsonData []byte, shopID, secretKey, tgUserID string) (string, error) {
	const op = "internal/bot/http_request/GetPaymentURL"

	// URL для запроса.
	url := "https://api.yookassa.ru/v3/payments"

	// Создание HTTP-запроса.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		slog.Error(op, "Ошибка создания HTTP-запроса", slog.String("error", err.Error()))
		return "", fmt.Errorf("%s: ошибка создания запроса: %w", op, err)
	}

	// Генерация уникального ключа для idempotence.
	idempotenceKey := fmt.Sprintf("payment-%d", time.Now().UnixNano())

	// Устанавливаем заголовки запроса.
	req.SetBasicAuth(shopID, secretKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotence-Key", idempotenceKey) // Уникальный ключ для повторных запросов.

	// Отправляем запрос.
	client := &http.Client{
		Timeout: 30 * time.Second, // Устанавливаем тайм-аут до 30 секунд.
	}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(op, "Ошибка выполнения запроса", slog.String("error", err.Error()))
		return "", err
	}
	defer resp.Body.Close()

	// Проверка статуса ответа.
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		// Логируем тело ответа для диагностики.
		body, _ := io.ReadAll(resp.Body)
		slog.Error(fmt.Sprintf("%s: неожиданный статус код: %d, тело ответа: %s", op, resp.StatusCode, string(body)))
		return "", fmt.Errorf("%s: неожиданный статус код: %d", op, resp.StatusCode)
	}

	// Декодируем ответ в структуру.
	var paymentResponse entities.PaymentResponse
	if err = json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		slog.Error(op, "Ошибка декодирования ответа", slog.String("error", err.Error()))
		return "", err
	}

	// Проверка наличия ссылки на оплату.
	if paymentResponse.Confirmation.ConfirmationURL == "" {
		err = fmt.Errorf("%s: не удалось получить ссылку на оплату", op)
		slog.Error(op, err)
		return "", err
	}

	// Возвращаем ссылку на оплату.
	return paymentResponse.Confirmation.ConfirmationURL, nil
}
