package entities

// PaymentResponse - структура ответааааа
type PaymentResponse struct {
	ID           string `json:"id"`
	Confirmation struct {
		ConfirmationURL string `json:"confirmation_url"`
	} `json:"confirmation"`
}
