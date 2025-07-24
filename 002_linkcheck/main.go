package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"linkcheck/model"
	"log"
	"net/http"
	"net/http/httputil"
	"net/smtp"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
)

var (
	siteUrl        string
	regFragment    *regexp.Regexp
	checkedUrlList []string            // 重複チェック防止
	errUrlList     []model.ErrorUrlMap // リンク切れだったURL
	// メール設定
	hostname string
	port     int
	username string
	password string
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	sitemapUrl := os.Getenv("SITEMAP_URL")
	siteUrl = os.Getenv("SITE_URL")

	reg, err := regexp.Compile(`^#`)
	if err != nil {
		log.Fatal(err)
		return
	}
	regFragment = reg

	sitemap, err := GetSitemap(sitemapUrl)
	if err != nil {
		log.Fatal("サイトマップの取得に失敗しました。")
		return
	}
	ParseSiteMap(sitemap)

	if len(errUrlList) > 0 {
		SendMail()
	} else {
		log.Println("リンク切れ無し")
	}
}

func GetSitemap(sitemapUrl string) (string, error) {
	req, err := http.NewRequest("GET", sitemapUrl, nil)
	if err != nil {
		log.Fatal("リクエストの作成に失敗しました")
		return "", err
	}
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("リクエストに失敗しました")
		return "", err
	}

	defer resp.Body.Close()
	byteArray, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("レスポンスの読み取りに失敗しました。")
		return "", err
	}

	return string(byteArray), nil
}

func ParseSiteMap(sitemapData string) {
	sitemap := model.Sitemap{}
	err := xml.Unmarshal([]byte(sitemapData), &sitemap)
	if err != nil {
		log.Fatal("サイトマップをパース出来ませんでした。")
		return
	}

	for _, item := range sitemap.UrlList {
		CheckURL(item.Location)
		time.Sleep(1 * time.Second)
	}
}

func CheckURL(targetUrl string) error {
	log.Println("Target:" + targetUrl)

	req, _ := http.NewRequest("GET", targetUrl, nil)
	client := new(http.Client)
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("url scraping failed")
		return err
	}

	var urlList []string
	doc.Find("a").Each(func(_ int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		urlParse, err := url.Parse(href)
		if err != nil {
			log.Fatal(err)
		}

		absoluteUrl := href
		// #はじまりならページ内リンクなので無視
		if regFragment.MatchString(href) {
			return
		}

		// 空なら無視
		if href == "" {
			return
		}

		// ドメイン無なら自サイトなので、絶対パスに変換
		if urlParse.Host == "" {
			absoluteUrl = siteUrl + href
		}

		// すでにチェック済みならスキップ
		if slices.Contains(checkedUrlList, absoluteUrl) {
			return
		}

		// リンク切れチェック
		log.Println("リンクチェック: " + absoluteUrl)

		req, _ := http.NewRequest("GET", absoluteUrl, nil)
		// カスタムヘッダー追加
		// req.Header.Set("Authorization", "Bearer access-token")

		client := new(http.Client)
		resp, err := client.Do(req)

		// 400番台以降はリンク切れとして扱う
		if resp.StatusCode > 400 {
			urlList = append(urlList, absoluteUrl)

			log.Println(
				fmt.Printf("リンク切れ StatusCode: %d", resp.StatusCode),
			)

			dumpResp, _ := httputil.DumpResponse(resp, true)
			fmt.Printf("%s \r\n", dumpResp)
		}

		checkedUrlList = append(checkedUrlList, absoluteUrl)
	})

	if len(urlList) > 0 {
		errUrlList = append(errUrlList, model.ErrorUrlMap{
			OriginUrl: targetUrl,
			UrlList:   urlList,
		})
	}

	return nil
}

// リンク切れをメールで通知
func SendMail() {
	hostname := os.Getenv("MAIL_HOST")
	port := os.Getenv("MAIL_PORT")

	from := os.Getenv("MAIL_FROM_ADDRESS")
	to := os.Getenv("MAIL_TO_ADDRESS")

	str := ""
	for _, errUrl := range errUrlList {
		str = str + "[エラーリンクのあるページURL]\n" + errUrl.OriginUrl + "\n[リンク切れURL]\n"
		for _, url := range errUrl.UrlList {
			str = str + "・" + url + "\n"
		}
	}

	msg := []byte(
		strings.ReplaceAll(
			fmt.Sprintf(
				"To: %s\nSubject: [Link Check]Error!\n\nリンク切れがあります。\n以下のページを修正してください。\n%s",
				to,
				str,
			),
			"\n",
			"\r\n",
		),
	)

	if err := smtp.SendMail(
		fmt.Sprintf("%s:%s", hostname, port),
		nil,
		from,
		[]string{to},
		msg,
	); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
