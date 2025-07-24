# linkcheck
サイトマップをもとにページを巡回してリンク切れを検知するスクリプト

[詳細はこちら](https://awatana.com/blog/page/17)

# how to use

```sh
$ cp .env.example .env
$ vi .env
SITEMAP_URL=https://example.com/sitemap.xml # サイトマップのURL
NOTICE_MAIL_ADDRESS=example@example.com # リンク切れ通知先のメールアドレス
```

# 注意
スクレイピングを行うのでよくよく動作確認してから使用してください。
(ページ数が多いサイトでの負荷等十分にチェックしていません。)
