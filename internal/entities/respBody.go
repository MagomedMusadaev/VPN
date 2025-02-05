package entities

type RespBodyKey struct {
	Key string `json:"accessUrl"`
}

func NewRespBody(key, status string) RespBodyKey {
	return RespBodyKey{
		Key: key,
	}
}
