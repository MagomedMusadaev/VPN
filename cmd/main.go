package main

import (
	"fmt"
	"log/slog"
	"os"

	"bot_vpn/internal/bot"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	const op = "cmd.main"

	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load("F:\\bot_vpn\\.env"); err != nil {
		slog.Error(op, "Ошибка загрузки .env файла", slog.String("error", err.Error()))
		return
	}

	// Получение токена бота
	tokenBot := os.Getenv("BOT_TOKEN")
	if tokenBot == "" {
		slog.Error(op, "Токен бота не найден в переменных окружения")
		return
	}

	// Инициализация API Telegram
	botAPI, err := tgbotapi.NewBotAPI(tokenBot)
	if err != nil {
		slog.Error(op, "Ошибка создания API бота", slog.String("error", err.Error()))
		return
	}

	botAPI.Debug = true
	slog.Info(op, fmt.Sprintf("Авторизован на аккаунте %s", botAPI.Self.UserName))

	// Настройка обновлений Telegram
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := botAPI.GetUpdatesChan(u)

	// Инициализация messenger и handler
	messenger := bot.NewMessengerBot(botAPI)
	handler := bot.NewCallbackHandler(messenger)

	// Основной цикл обработки обновлений
	for update := range updates {
		switch {
		case update.Message != nil && update.Message.Text != "":
			handler.Message(update)
		case update.CallbackQuery != nil:
			handler.Button(update.CallbackQuery)
		}
	}
}
