# コミットメッセージのルール

## 形式

```text
#<Issue 番号> <type>: <subject> (<スコープ>)

<本文>

<トレーラ>
```

例:

```text
#722 fix: E2E の起動待ち先とカバレッジ集計範囲を是正 (frontend)

webServer.url が admin に存在しない /about を指しており、Playwright の起動判定
(200 以上 404 未満) を満たさず必ずタイムアウトしていた。未認証かつ Admin API
未起動でも 200 を返す /signin に変更した。あわせて初回コンパイルが重い実態に合わせ
timeout を 120 秒に明示した。

collectCoverageFrom が components / lib / app の 3 つしか列挙しておらず、ソースが
集中する features 配下が集計対象外だった。実装を持つ 7 ディレクトリに広げ、型定義
のみのディレクトリと再エクスポートのみの index.ts を除外した。

Co-Authored-By: Claude Opus 5 (1M context) <noreply@anthropic.com>
```

| 要素       | 必須 | 内容                                                        |
| ---------- | ---- | ----------------------------------------------------------- |
| Issue 番号 | 必須 | `#722` の形式。対応する Issue が無い場合は先に Issue を作る |
| type       | 必須 | `commit-types.md` の type                                   |
| subject    | 必須 | 日本語で 1 行                                               |
| スコープ   | 必須 | `labels.md` のラベルを丸括弧で末尾に置く                    |

## subject

- 日本語で書く
- 体言止めまたは「〜する」形で端的に記述する
- 50 文字以内を目安にする
- 末尾に句点をつけない

## 本文

**必須。** subject の後に空行を 1 行あけて書く。

- **何が問題だったか → どう変えたか** の順で書く。変更の理由が subject から自明でも省かない
- 1 段落 1 論点にし、論点が変われば空行で段落を分ける
- 数値・設定値・対象範囲など、判断の根拠になった実測値を書く
- 長い行は 80 文字程度で折り返す
- subject と違い、本文には句点を使う

git diff を見ればわかること (変更したファイル名、コードそのもの) は書かない。

## 例

```text
#12 feat: 優先パスの申込ハンドラを追加 (backend)
#20 docs: README に開発環境の起動手順を追記 (docs)
#31 chore: Firestore エミュレータを docker compose に追加 (repo)
```
