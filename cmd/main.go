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
)

func main() {
	const op = "cmd.main"

	// Загрузка переменных окружения из .env файла
	if err := godotenv.Load("../.env"); err != nil {
		slog.Error(op, "Ошибка загрузки .env файла", slog.String("error", err.Error()))
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

	connDB, err := bot.ConnectPostgresDB() // Подключение к базе данных PostgreSQL
	if err != nil {
		os.Exit(1)
	}
	//client, err := bot.NewRedis() // Подключение Redis
	//if err != nil {
	//	os.Exit(1)
	//}

	// Создание зависимостей
	//repo := bot.NewRepo(connDB, client)
	repo := bot.NewRepo(connDB)
	messenger := bot.NewMessengerBot(botAPI, bot.NewHttpRequest(), bot.NewKeyBoard(), repo)
	httpHandler := bot.NewHttpHandler(messenger)
	handler := bot.NewCallbackHandler(messenger)

	//Инициализация маршрутов HTTP
	bot.InitRout(httpHandler)

	//Запуск HTTP-сервера
	go func() {
		slog.Info("Сервер запущен", slog.String("url", "http://localhost:"+port))
		if err = http.ListenAndServe(":"+port, nil); err != nil {
			slog.Error(op, "Ошибка запуска сервера", slog.String("error", err.Error()))
			os.Exit(1) // Завершаем приложение, если сервер упал
		}
	}()

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

//TODO
// ХРАНИМ:
// chat_id BIGINT UNIQUE NOT NULL,            -- Уникальный идентификатор пользователя в Telegram
// user_tg_id BIGINT NOT NULL,               -- Telegram ID пользователя (не всегда совпадает с chat_id, если в группе)
// НЕ ХРАНИИМ:
// id SERIAL PRIMARY KEY,                    -- Уникальный идентификатор записи
// key_id INT UNIQUE NOT NULL,                -- ID ключа в системе Outline Manager
// key VARCHAR(255) NOT NULL,                 -- Ключ подключения (например, токен или идентификатор)
// key_status VARCHAR(50) NOT NULL,           -- Статус ключа (например, активен, истёк, ограничен)
// key_start_date TIMESTAMP NOT NULL,        -- Дата начала действия ключа
// key_expiry_date TIMESTAMP NOT NULL,       -- Дата истечения действия ключа
// created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP -- Дата первой записи данных пользователя

//Таблица users (Пользователи)
//Поле						Тип данных									Описание
//id						SERIAL (Primary Key)						Уникальный ID пользователя в базе
//telegram_id				BIGINT UNIQUE								ID пользователя в Telegram
//chat_id					BIGINT UNIQUE								ID чата с пользователем
//referral_code 			VARCHAR(255) UNIQUE							Реферальная ссылка пользователя
//created_at				TIMESTAMP DEFAULT CURRENT_TIMESTAMP			Время создания записи
//referred_by				INT											ID пользователя кто его пригласил
//FOREIGN KEY (referred_by) REFERENCES users(id) ON DELETE SET NULL  	Внешний ключ на приглашённого

//Таблица keys (Ключи доступа)
//Поле						Тип данных									Описание
//id						SERIAL (Primary Key)						Уникальный ID ключа
//user_id					INT (Foreign Key)							Ссылка на пользователя (users.id)
//key						VARCHAR(255)								Ключ доступа
//created_at				TIMESTAMP									Время создания ключа
//expires_at				TIMESTAMP									Время истечения ключа
//CHECK (expires_at > created_at)         								Проверка, что время истечения ключа больше времени создания

//CREATE INDEX idx_referral_code ON users(referral_code);
//CREATE INDEX idx_telegram_id ON users(telegram_id);
//CREATE INDEX idx_referred_by ON users(referred_by); - В ДАЛЁКОМ БУДУШЕМ (ЕСЛИ ПРОЕКТ БУДЕТ ЖИТЬ)
