package entities

// PaymentRequest представляет тело запроса на создание платежаа
type PaymentRequest struct {
	Capture      bool              `json:"capture"`      // Флаг захвата платежа
	Description  string            `json:"description"`  // Описание платежа
	Amount       Amount            `json:"amount"`       // Сумма платежа
	Confirmation Confirmation      `json:"confirmation"` // Параметры подтверждения
	Metadata     map[string]string `json:"metadata"`     // Метаданные
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
