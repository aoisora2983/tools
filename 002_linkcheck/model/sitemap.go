package model

import (
	"encoding/xml"
)

type Sitemap struct {
	XMLName xml.Name  `xml:"urlset"`
	UrlList []UrlItem `xml:"url"`
	Xmlns   string    `xml:"xmlns,attr"`
}

type UrlItem struct {
	Location string `xml:"loc"`
}
