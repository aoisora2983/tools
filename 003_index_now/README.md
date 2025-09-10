# IndexNow
サイトマップをもとにIndexNowにページの追加・削除を通知するスクリプト

[詳細はこちら](https://awatana.com/blog/page/30)

# how to use

```sh
$ cp .env.example .env
$ vi .env
SITEMAP_URL=https://example.com/sitemap.xml # サイトマップのURL
INDEXNOW_API_URL=https://api.indexnow.org/IndexNow

HOST="example.com" # Webmasterツールを登録したサイトのホスト
KEY="[APIキー]" # IndexNowで使用するAPIキー
KEY_LOCATION="https://example.com/[APIキー].txt" # APIキーファイルを配置した場所

$ go build main.go
$ ./main
```
