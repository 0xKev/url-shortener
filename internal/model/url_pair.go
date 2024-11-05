package model

type URLPair struct {
	ShortSuffix string `json:"shortSuffix"`
	BaseURL     string `json:"baseURL"`
	Domain      string `json:"domain"`
	Error       string `json:"error"`
}
