package entities

type MetaData struct {
	Metadata struct {
		UserID string `json:"user_id"`
		Tariff string `json:"tariff"`
	} `json:"metadata"`
}

func NewMetaData(userId, tariff string) *MetaData {
	return &MetaData{
		Metadata: struct {
			UserID string `json:"user_id"`
			Tariff string `json:"tariff"`
		}{UserID: userId, Tariff: tariff},
	}
}
