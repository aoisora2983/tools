package model

type CheckItem struct {
	URL     string // チェック対象URL
	Checked bool   // チェック済みか
}

func NewCheckItem(url string) *CheckItem {
	check := new(CheckItem)
	check.URL = url
	check.Checked = false
	return check
}
