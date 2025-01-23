package bot

import (
	"bot_vpn/internal/entities"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type HttpRequestInt interface {
	SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, error)
}

type HttpRequest struct {
}

func NewHttpRequest() *HttpRequest {
	return &HttpRequest{}
}

// createInsecureHTTPClient - создает и возвращает HTTP-клиент с отключенной проверкой SSL. TODO: избавиться от этого в конце
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
func (h *HttpRequest) SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, error) { // TODO: нужно дорабоать (возможно)
	const op = "internal/bot/http_handler/SendKeyRequest"

	// Сериализация полезной нагрузки в JSON.
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error(op, "Ошибка сериализации полезной нагрузки", slog.String("error", err.Error()))
		return "", err
	}

	// Создание нового HTTP-запроса.
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error(op, "Ошибка создания HTTP-запроса", slog.String("error", err.Error()))
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	// Используем HTTP-клиент с отключенной проверкой SSL.
	client := createInsecureHTTPClient()

	// Выполнение HTTP-запроса.
	resp, err := client.Do(req)
	if err != nil {
		slog.Error(op, "Ошибка выполнения HTTP-запроса", slog.String("error", err.Error()))
		return "", err
	}
	defer resp.Body.Close()

	// Проверка статуса ответа.
	if resp.StatusCode != http.StatusCreated {
		slog.Error(op, "Неверный статус ответа", slog.Int("status_code", resp.StatusCode))
		return "", fmt.Errorf("получен статус %d, ожидался %d", resp.StatusCode, http.StatusCreated)
	}

	// Распаковка ответа.
	var respBody entities.RespBody
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.Error(op, "Ошибка декодирования тела ответа", slog.String("error", err.Error()))
		return "", err
	}

	// Возвращаем ключ из ответа.
	return respBody.Key, nil
}
