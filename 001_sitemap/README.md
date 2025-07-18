# sitemap
サイトのルートからリンクを辿ってサイトマップを作成するgolangスクリプト

[詳細はこちら](https://awatana.com/blog/page/15)

# how to use

```sh
$ cp .env.example .env
$ vi .env
URL=https://example.com # サイトマップを作成したい任意のサイトのドメインに変更
```

# 注意
スクレイピングを行うのでよくよく動作確認してから使用してください。
(ページ数が多いサイトでの負荷や、
複雑なサイト構成の場合にループが発生しないか等十分にチェックしていません。)
