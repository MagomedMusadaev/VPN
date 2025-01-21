package entities

type KeyPayload struct {
	UserID    int64  `json:"user_id"`
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
}

func NewKeyPayload(userID int64, userName, firstName string) KeyPayload {
	return KeyPayload{
		UserID:    userID,
		UserName:  userName,
		FirstName: firstName,
	}
}
