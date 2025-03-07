package entities

// PaymentRequest представляет тело запроса на создание платежа
type PaymentRequest struct {
	Capture      bool              `json:"capture"`      // Флаг захвата платежа
	Description  string            `json:"description"`  // Описание платежа
	Amount       Amount            `json:"amount"`       // Сумма платежа
	Confirmation Confirmation      `json:"confirmation"` // Параметры подтверждения
	Metadata     map[string]string `json:"metadata"`     // Метаданные
	Receipt      Receipt           `json:"receipt"`      // Данные чека (обязательны для YooKassa)
}

// Amount представляет сумму платежа
type Amount struct {
	Value    string `json:"value"`    // Сумма
	Currency string `json:"currency"` // Валюта (например, "RUB")
}

// Confirmation представляет параметры подтверждения платежа
type Confirmation struct {
	Type      string `json:"type"`       // Тип подтверждения (например, "redirect")
	ReturnURL string `json:"return_url"` // URL для редиректа после оплаты
}

// Receipt представляет информацию о чеке для YooKassa
type Receipt struct {
	Email string        `json:"email"` // Email покупателя (обязателен, если нет телефона)
	Items []ReceiptItem `json:"items"` // Список товаров/услуг
}

// ReceiptItem представляет отдельную позицию в чеке
type ReceiptItem struct {
	Description    string `json:"description"`     // Описание товара/услуги
	Quantity       int    `json:"quantity"`        // Количество
	Amount         Amount `json:"amount"`          // Цена за единицу
	VATCode        int    `json:"vat_code"`        // Код НДС (1 - 20%, 2 - 10%, 6 - без НДС)
	PaymentMode    string `json:"payment_mode"`    // Способ оплаты ("full_payment" — полная оплата)
	PaymentSubject string `json:"payment_subject"` // Тип товара ("service" — услуга)
}
