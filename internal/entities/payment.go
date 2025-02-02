package entities

// Event представляет данные уведомления от ЮКасса.
type Event struct {
	ID            string                 `json:"id"`
	Status        string                 `json:"status"`
	Paid          bool                   `json:"paid"`
	Amount        AmountResp             `json:"amount"`
	CreatedAt     string                 `json:"created_at"`
	Description   string                 `json:"description"`
	Metadata      map[string]interface{} `json:"metadata"`
	PaymentMethod PaymentMethod          `json:"payment_method"`
}

// AmountResp - представляет сумму платежа.
type AmountResp struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// PaymentMethod содержит информацию о методе платежа.
type PaymentMethod struct {
	Type  string `json:"type"` // Например, "bank_card"
	ID    string `json:"id"`
	Saved bool   `json:"saved"`
	Card  Card   `json:"card"`
	Title string `json:"title"`
}

// Card содержит информацию о карте, использованной для платежа.
type Card struct {
	First6        string      `json:"first6"`
	Last4         string      `json:"last4"`
	ExpiryMonth   string      `json:"expiry_month"`
	ExpiryYear    string      `json:"expiry_year"`
	CardType      string      `json:"card_type"`
	CardProduct   CardProduct `json:"card_product"`
	IssuerCountry string      `json:"issuer_country"`
	IssuerName    string      `json:"issuer_name"`
}

// CardProduct содержит информацию о продукте карты.
type CardProduct struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// Recipient содержит информацию о получателе.
type Recipient struct {
	AccountID string `json:"account_id"`
	GatewayID string `json:"gateway_id"`
}
