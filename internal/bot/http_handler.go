package bot

import (
	"bot_vpn/internal/entities"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type HttpHandlerInt interface {
	SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, error)
}

type HttpHandler struct {
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

func createInsecureHTTPClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Отключаем проверку SSL
		},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Minute, // Устанавливаем тайм-аут
	}
	return client
}

func (t *HttpHandler) SendKeyRequest(apiURL string, payload entities.KeyPayload) (string, error) {
	const op = "internal/bot/http_handler"

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error(op, err)
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		slog.Error(op, err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := createInsecureHTTPClient()

	resp, err := client.Do(req)
	if err != nil {
		slog.Error(op, err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error(op, resp.StatusCode) // TODO: доработать 201
		//return "", err
	}

	var respBody entities.RespBody
	if err = json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		slog.Error(op, err)
		return "", err
	}

	slog.Info(op, "key generated successfully:", respBody.Key)

	return respBody.Key, nil
}

// TODO:
// разобраться с логами
// разобраться с названиями ключей
//
