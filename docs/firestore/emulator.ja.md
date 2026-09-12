# Firestore エミュレータ

[English](./emulator.md) | [日本語](./emulator.ja.md)

[ドキュメント一覧](../README.ja.md)に戻る。

ローカル開発では Firebase Emulator Suite の Firestore エミュレータを使います。
コンテナの中で動くため、手元に JDK と Firebase CLI を入れる必要はありません。

## 起動と停止

```sh
just up
just down
```

`just up` はエミュレータが healthy になるまで待ってから戻ります。初回はイメージのビルドが走ります。JRE と Firebase CLI を入れるため数分かかります。

## ポート

| ポート  | 用途                                                   |
| ------- | ------------------------------------------------------ |
| `18080` | Firestore エミュレータ。`8080` 以降は api サーバに残す |
| `4000`  | Emulator UI                                            |
| `4400`  | Emulator hub                                           |
| `4500`  | UI が読むログの配信                                    |
| `9150`  | UI が変化を受け取る websocket                          |

UI は <http://localhost:4000/firestore> で開きます。

## クライアントからの接続

次の環境変数を設定すると、SDK は Google Cloud ではなくエミュレータへ接続します。

```sh
export FIRESTORE_EMULATOR_HOST=localhost:18080
```

ローカルのプロジェクト ID は `demo-aozora-park` です。
`demo-` で始まる ID は SDK が本物の Google Cloud へ接続することを拒むため、
設定を誤ったクライアントが本番のデータに触れる経路がありません。

## api からの接続

`backend/justfile` が必要なレシピにだけ環境変数を渡すため、追加の設定は要りません。

```sh
cd backend
just run-api  # エミュレータに向けて api を起動する
```

`GOOGLE_CLOUD_PROJECT` にコードの既定値はありません。
設定が無いと api は起動を拒みます。
誤った設定のまま別のプロジェクトへ繋ぐより、起動の時点で失敗するほうが安全なためです。

## テストは別のエミュレータを使う

このコンテナはアプリを手で動かすためのものです。テストはこれを使いません。

```sh
cd backend
just test           # エミュレータが要るテストは飛ばされる
just test-emulator  # パッケージごとにコンテナを起動して実行する
```

テストは testcontainers で自分でエミュレータを起動し、`DOCKER_TESTS` で入り切りを決めます。
環境変数が無ければ自分で飛ばすため、Docker が無い環境でも `just test` は通ります。
2 つを分けておくと、UI で見ているデータをテストが読んだり消したりすることがありません。

## 読み書きの確認

```sh
BASE="http://localhost:18080/v1/projects/demo-aozora-park/databases/(default)/documents"

curl -X POST "$BASE/parks?documentId=aozora" \
  -H 'Content-Type: application/json' \
  -d '{"fields":{"name":{"stringValue":"Aozora Park"}}}'

curl "$BASE/parks/aozora"
```

データはメモリ上にだけ置かれます。コンテナを止めるとすべて消えます。
