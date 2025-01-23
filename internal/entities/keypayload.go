package entities

type KeyPayload struct {
	UserID string `json:"name"`
}

func NewKeyPayload(userID string) KeyPayload {
	return KeyPayload{
		UserID: userID,
	}
}
