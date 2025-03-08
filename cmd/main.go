package main

import (
	"bot_vpn/internal/bot"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log/slog"
	"net/http"
	"os"
	"time"
)

func main() {
	const op = "cmd/main"

	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load("F:\\bot_vpn\\.env"); err != nil {
		slog.Warn(op, "Ошибка загрузки .env файла", slog.String("error", err.Error()))
	}

	port := os.Getenv("PORT")
	if port == "" {
		slog.Error(op, "Не указан порт для сервера")
		return
	}

	// Инициализация API Telegram и получение обновлений
	botAPI, updates, err := initBotAPI()
	if err != nil {
		slog.Error(op, "Ошибка инициализации бота", err.Error())
		return
	}

	// Отключение режима отладки (true - чтобы заработало)
	botAPI.Debug = false
	slog.Info(fmt.Sprintf("Авторизован на аккаунте %s", botAPI.Self.UserName))

	connDB, err := bot.ConnectPostgresDB() // Подключение к базе данных Psql
	if err != nil {
		os.Exit(1)
	}
	client, err := bot.NewRedis() // Подключение Redis
	if err != nil {
		os.Exit(1)
	}

	// Создание зависимостей
	repo := bot.NewRepo(connDB, client)
	messenger := bot.NewMessengerBot(botAPI, bot.NewHttpRequest(), bot.NewKeyBoard(), repo, client)
	httpHandler := bot.NewHttpHandler(messenger)
	handler := bot.NewCallbackHandler(messenger)

	go messenger.SetLimitForExpiredKeys()

	// Инициализация маршрутов HTTP
	bot.InitRout(httpHandler)

	// Запуск HTTP-сервера
	go func() {
		slog.Info("Сервер запущен", slog.String("url", "http://localhost:"+port))
		if err = http.ListenAndServe(":"+port, nil); err != nil {
			slog.Error(op, "Ошибка запуска сервера", slog.String("error", err.Error()))
			os.Exit(1) // Завершаем приложение, если сервер упал
		}
	}()

	// Устанавливаем часовой пояс для всего приложения
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		slog.Error(op, "Ошибка загрузки часового пояса:", slog.String("error", err.Error()))
		return
	}

	// Устанавливаем локальный часовой пояс по умолчанию
	time.Local = loc
	slog.Info("Текущее время по МСК", slog.Time("time", time.Now()))

	// Основной цикл обработки обновлений Telegram
	for update := range updates {
		switch {
		// Если пришло текстовое сообщение
		case update.Message != nil && update.Message.Text != "":
			handler.Message(update) // Обработка текстового сообщения
		// Если пришёл callback (например, кнопка)
		case update.CallbackQuery != nil:
			handler.Button(update.CallbackQuery) // Обработка callback
		}
	}
}

// initBotAPI - функция для инициализации API Telegram и получения обновлений
func initBotAPI() (*tgbotapi.BotAPI, tgbotapi.UpdatesChannel, error) {
	// Получение токена бота из переменных окружения
	tokenBot := os.Getenv("BOT_TOKEN")
	if tokenBot == "" {
		return nil, nil, fmt.Errorf("отсутствует токен бота")
	}

	// Инициализация API Telegram
	botAPI, err := tgbotapi.NewBotAPI(tokenBot)
	if err != nil {
		return nil, nil, err
	}

	// Настройка получения обновлений от Telegram
	u := tgbotapi.NewUpdate(0)          // Установка сдвига обновлений
	u.Timeout = 60                      // Тайм-аут для ожидания новых обновлений
	updates := botAPI.GetUpdatesChan(u) // Канал для получения обновлений

	return botAPI, updates, nil
}
