package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"indexnow/model"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	sitemapUrl := os.Getenv("SITEMAP_URL")
	indexNowApiUrl := os.Getenv("INDEXNOW_API_URL")
	host := os.Getenv("HOST")
	key := os.Getenv("KEY")
	keyLocation := os.Getenv("KEY_LOCATION")

	// 1日分前のサイトマップを取得
	oldSitemap, err := GetSitemapByFile("./work/sitemap.xml")

	// 本日分サイトマップを取得
	newSitemap, err := GetSitemapByUrl(sitemapUrl)

	// 本日分情報をworkに退避・上書き
	file, err := os.Create("./work/sitemap.xml")
	file.Write(newSitemap)

	if err != nil {
		log.Fatal("サイトマップの取得に失敗しました。")
		return
	}

	oldUrlList := ParseSiteMap(oldSitemap)
	newUrlList := ParseSiteMap(newSitemap)

	// 1日前のデータと本日分を比較
	urlList := GetDiffUrl(oldUrlList, newUrlList)

	if len(urlList) == 0 {
		log.Println("通知が必要なURLはありません。")
		return
	}

	// APIを実行する
	PostIndexNow(urlList, indexNowApiUrl, host, key, keyLocation)
}

func GetSitemapByFile(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		fmt.Println("ファイルを開けませんでした : " + filepath)
		return nil, nil
	}

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("ファイルを読み取れませんでした。 : " + filepath)
		return nil, nil
	}

	file.Close()

	return data, nil
}

func GetSitemapByUrl(sitemapUrl string) ([]byte, error) {
	req, err := http.NewRequest("GET", sitemapUrl, nil)
	if err != nil {
		log.Fatal("リクエストの作成に失敗しました")
		return nil, err
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("リクエストに失敗しました")
		return nil, err
	}

	defer resp.Body.Close()
	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("レスポンスの読み取りに失敗しました。")
		return nil, err
	}

	return byteArray, nil
}

func ParseSiteMap(sitemapData []byte) []string {
	var urlList []string
	sitemap := model.Sitemap{}

	if sitemapData == nil {
		return urlList
	}

	err := xml.Unmarshal(sitemapData, &sitemap)
	if err != nil {
		log.Fatal("サイトマップをパース出来ませんでした。")
	}

	for _, item := range sitemap.UrlList {
		// 書き込む
		urlList = append(urlList, item.Location)
	}

	return urlList
}

func GetDiffUrl(oldUrlList []string, newUrlList []string) []string {
	var diffUrlList []string
	diffMap := map[string]int{}

	for _, oldUrl := range oldUrlList {
		diffMap[oldUrl] = 1
	}

	for _, newUrl := range newUrlList {
		_, ok := diffMap[newUrl]

		if ok {
			diffMap[newUrl]++
		} else {
			diffMap[newUrl] = 1
		}
	}

	for url, cnt := range diffMap {
		if cnt == 1 {
			diffUrlList = append(diffUrlList, url)
		}
	}

	return diffUrlList
}

func PostIndexNow(
	urlList []string,
	apiUrl string,
	host string,
	key string,
	keyLocation string,
) error {
	postData := model.IndexNowData{
		Host:        host,
		Key:         key,
		KeyLocation: keyLocation,
		UrlList:     urlList,
	}
	jsonData, err := json.Marshal(postData)
	if err != nil {
		log.Fatal("リクエストデータの作成に失敗しました")
		return err
	}

	log.Println("URL:" + apiUrl)
	log.Println("DATA:" + string(jsonData))

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatal("リクエストの作成に失敗しました")
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("リクエストに失敗しました")
		return err
	}

	defer resp.Body.Close()

	// 成功したら抜ける
	if resp.Status == "200 OK" {
		log.Println("リクエストに成功しました。")
		return nil
	}

	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("レスポンスの読み取りに失敗しました。")
		return err
	}

	var response model.Response
	if err := json.Unmarshal(byteArray, &response); err != nil {
		log.Println(string(byteArray))
		log.Fatal("レスポンスのJson化に失敗しました")
		return err
	}

	// エラーの詳細をログに出力しておく
	log.Println(response.Code)
	log.Println(response.Message)

	for _, detail := range response.Details {
		log.Println(detail.Target)
		log.Println(detail.Message)
	}

	return nil
}
