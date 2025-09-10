package model

type IndexNowData struct {
	Host        string   `json:"host"`
	Key         string   `json:"key"`
	KeyLocation string   `json:"keyLocation"`
	UrlList     []string `json:"UrlList"`
}
