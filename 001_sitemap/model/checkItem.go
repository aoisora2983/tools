package model

type CheckItem struct {
	URL       string // チェック対象URL
	Checked   bool   // チェック済みか
	IsNoindex bool   // noindexが設定されているURLか
}

func NewCheckItem(url string) *CheckItem {
	check := new(CheckItem)
	check.URL = url
	check.Checked = false
	check.IsNoindex = false
	return check
}
