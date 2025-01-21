package entities

type RespBody struct {
	Key string `json:"accessUrl"`
}

func NewRespBody(key, status string) RespBody {
	return RespBody{
		Key: key,
	}
}
