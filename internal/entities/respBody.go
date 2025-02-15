package entities

type RespBodyKey struct {
	Key string `json:"accessUrl"`
	ID  string `json:"id"`
}

func NewRespBody(key, id string) RespBodyKey {
	return RespBodyKey{
		Key: key,
		ID:  id,
	}
}
