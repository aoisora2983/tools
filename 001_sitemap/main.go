package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sitemap/creator/model"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

var checkList = make(map[string]*model.CheckItem)
var siteUrl = ""
var xmlns = "http://www.sitemaps.org/schemas/sitemap/0.9"

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	formatUrl, err := formatURL(os.Getenv("URL"))
	siteUrl = formatUrl
	log.Println("Target サイト: " + formatUrl)
	// ルートURLは最初に追加しておく
	checkList[formatUrl] = model.NewCheckItem(formatUrl)
	nestUrl(formatUrl)
	createSiteMap()
}

// urlを辿る
func nestUrl(targetUrl string) error {
	// 閲覧済みフラグをON
	checkList[targetUrl].Checked = true

	log.Println("Target:" + targetUrl)

	time.Sleep(1 * time.Second)

	req, _ := http.NewRequest("GET", targetUrl, nil)
	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("url scraping failed")
		return err
	}

	siteUrlParse, err := url.Parse(siteUrl)
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")

		urlParse, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}

		absoluteUrl := ""
		// ドメイン無なら自サイトなので、絶対パスに変換
		if urlParse.Host == "" {
			absoluteUrl = siteUrl + href
		} else {
			// ドメイン有ならチェック
			if urlParse.Host == siteUrlParse.Host {
				// 自サイトドメインなら追加
				absoluteUrl = href
			} else {
				// 他サイトドメインならスキップ
				return
			}
		}

		formatUrl, err := formatURL(absoluteUrl)
		if err != nil {
			log.Fatal(err)
		}

		_, hasUrl := checkList[formatUrl] // 存在チェック
		if !hasUrl {                      // 存在しない物のみ追加
			checkList[formatUrl] = model.NewCheckItem(formatUrl)
		}
	})

	// チェックリストが全てOKになったら抜ける
	allOk := true
	for _, checkItem := range checkList {
		if !checkItem.Checked {
			allOk = false
		}
	}
	if allOk {
		return nil
	}

	for checkUrl, checkItem := range checkList {
		// 既に閲覧済みならスキップ
		if checkItem.Checked {
			continue
		} else {
			nestUrl(checkUrl)
		}
	}

	return nil
}

// サイトマップ作製
func createSiteMap() {
	sitemap := &model.Sitemap{
		Xmlns: xmlns,
	}

	for _, item := range checkList {
		sitemap.UrlList =
			append(
				sitemap.UrlList,
				model.UrlItem{
					Location: item.URL,
				},
			)
	}

	output, err := xml.MarshalIndent(sitemap, "", "    ")
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	f, err := os.Create("output/sitemap.xml")
	f.Write([]byte(xml.Header))
	f.Write(output)
}

// URL整形
func formatURL(absoluteUrl string) (string, error) {
	// 末尾パラメータ除去
	parseUrl, err := url.Parse(absoluteUrl)
	if err != nil {
		return "", err
	}

	// 末尾「/」除去
	formatUrl := parseUrl.Scheme + "://" + parseUrl.Hostname() + parseUrl.Path
	formatUrl = strings.TrimSuffix(formatUrl, "/")

	return formatUrl, nil
}
