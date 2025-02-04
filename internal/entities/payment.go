package entities

// AmountResp - структура для хранения информации о сумме и валюте.
type AmountResp struct {
	Value    string `json:"value"`    // Значение суммы (например, "100.00")
	Currency string `json:"currency"` // Валюта суммы (например, "RUB")
}

// CardProduct - структура для хранения информации о карточном продукте.
type CardProduct struct {
	Code string `json:"code"` // Код карточного продукта (например, "E")
}

// Card - структура для хранения информации о платёжной карте.
type Card struct {
	First6        string      `json:"first6"`         // Первые 6 цифр карты
	Last4         string      `json:"last4"`          // Последние 4 цифры карты
	ExpiryYear    string      `json:"expiry_year"`    // Год истечения срока действия карты
	ExpiryMonth   string      `json:"expiry_month"`   // Месяц истечения срока действия карты
	CardType      string      `json:"card_type"`      // Тип карты (например, "MasterCard")
	CardProduct   CardProduct `json:"card_product"`   // Информация о карточном продукте
	IssuerCountry string      `json:"issuer_country"` // Страна эмитента карты (например, "US")
}

// PaymentMethod - структура для хранения информации о платёжном методе.
type PaymentMethod struct {
	Type   string `json:"type"`   // Тип платёжного метода (например, "bank_card")
	ID     string `json:"id"`     // Уникальный ID платёжного метода
	Saved  bool   `json:"saved"`  // Флаг, указывающий, сохранён ли метод
	Status string `json:"status"` // Статус платёжного метода (например, "inactive")
	Title  string `json:"title"`  // Название платёжного метода (например, "Bank card *4477")
	Card   Card   `json:"card"`   // Информация о карте, если платёжный метод - карта
}

// AuthorizationDetails - структура для хранения информации о деталях авторизации.
type AuthorizationDetails struct {
	RRN          string       `json:"rrn"`            // Резервный номер транзакции (RRN)
	AuthCode     string       `json:"auth_code"`      // Код авторизации транзакции
	ThreeDSecure ThreeDSecure `json:"three_d_secure"` // Информация о 3D Secure
}

// ThreeDSecure - структура для хранения данных о протоколе 3D Secure.
type ThreeDSecure struct {
	Applied            bool   `json:"applied"`             // Применён ли протокол 3D Secure (true/false)
	Protocol           string `json:"protocol"`            // Протокол 3D Secure (например, "v1")
	MethodCompleted    bool   `json:"method_completed"`    // Завершён ли метод 3D Secure
	ChallengeCompleted bool   `json:"challenge_completed"` // Пройден ли challenge для 3D Secure
}

// Object - структура для хранения информации о платеже.
type Object struct {
	ID                   string               `json:"id"`                    // Уникальный идентификатор платёжной транзакции
	Status               string               `json:"status"`                // Статус платёжной транзакции (например, "succeeded")
	Amount               Amount               `json:"amount"`                // Сумма платежа
	IncomeAmount         Amount               `json:"income_amount"`         // Сумма, полученная после вычета комиссии
	Description          string               `json:"description"`           // Описание транзакции (например, "Оплата подписки на 1 месяц")
	Recipient            map[string]string    `json:"recipient"`             // Информация о получателе
	PaymentMethod        PaymentMethod        `json:"payment_method"`        // Информация о платёжном методе
	CapturedAt           string               `json:"captured_at"`           // Время захвата платежа
	CreatedAt            string               `json:"created_at"`            // Время создания транзакции
	Test                 bool                 `json:"test"`                  // Признак тестовой транзакции
	RefundedAmount       Amount               `json:"refunded_amount"`       // Сумма возврата, если транзакция была отменена
	Paid                 bool                 `json:"paid"`                  // Был ли платёж выполнен
	Refundable           bool                 `json:"refundable"`            // Флаг, указывающий, можно ли вернуть деньги
	Metadata             map[string]string    `json:"metadata"`              // Метаданные транзакции (например, идентификатор пользователя)
	AuthorizationDetails AuthorizationDetails `json:"authorization_details"` // Детали авторизации
}

// Event - структура для хранения данных о событии (например, успешной платёжной транзакции).
type Event struct {
	Event  string `json:"event"`  // Название события (например, "payment.succeeded")
	Object Object `json:"object"` // Объект, связанный с событием (данные транзакции)
}
