# backend のパッケージ構成

[English](./README.md) | [日本語](./README.ja.md)

[ドキュメント一覧](../../README.ja.md)に戻る。

`backend/` の内部構造をまとめます。規約は `.claude/rules/go/coding.md` が持ちます。ここには現状と、その形を選んだ理由を置きます。

## 層と依存の向き

外側をビジネス機能で分け、機能の内側を層で分けます。依存は内側に向かうものだけを許します。

![層と依存の向き](./images/layers.png)

| リング                     | パッケージ                                                                                              |
| -------------------------- | ------------------------------------------------------------------------------------------------------- |
| Enterprise Business Rules  | `park/domain/model`、`inventory/domain/model`、`platform/serving/apperr`                                |
| Application Business Rules | `park/usecase`、`park/domain/repository`、`inventory/usecase`、`inventory/domain/repository`            |
| Interface Adapters         | `<機能>/handler`、`<機能>/infrastructure/firestore`、`internal/park` と `internal/inventory` (組み立て) |
| Frameworks & Drivers       | `cmd/api`、`internal/platform/**`、`gen/**`                                                             |

## パッケージ間の依存

矢印は import の向きです。`go list` の出力から起こしています。

![パッケージ依存](./images/dependencies.png)

現在の依存を引くコマンドはこれです。

```sh
cd backend
go list -f '{{range .Imports}}{{.}}{{end}}' ./internal/park/usecase   # このパッケージが知っているもの
go list -f '{{.ImportPath}} {{.Imports}}' ./... | grep park/domain/model  # このパッケージを知っているもの
```

## apperr だけ置き場所とリングがずれる

`platform/serving/apperr` はディレクトリでは外側の `internal/platform/` にありますが、依存の向きでは中心にあります。`domain/model` がセンチネルエラーを生成するのに使うためです。

外を向いた知識は `Kind` から `connect.Code` への変換メソッドだけで、それを呼ぶのは外側の `interceptor` です。図では中心に破線で置いています。

## 判断の記録

**`usecase` にインタフェースを置かない。** `handler` から `usecase` への依存は最初から内側を向いているので、反転させるものがありません。1 つの取り決めに対する実装も 1 つだけです。Go は実装側に宣言が要らないため、必要になった時点で `handler` 側にインタフェースを切れます。切る時点は、同じ取り決めに 2 つ目の実装が要るとき (キャッシュ、機能フラグ) か、`handler` のテストで `usecase` ごと差し替えたくなったときです。

**`infrastructure/firestore` は技術名のままにする。** 抽象名は `domain/repository` の `Reader` と `Writer` が持ちます。実装側を `datastore` のような抽象名にすると、2 つ目の実装を追加するときに名前が空きません。役割は型名 (`firestore.Repository`) で表します。

**`park` が値オブジェクトにしたのは `ParkID` だけ。** 表示名と上限人数と日数は `Park` の外へ単体で出ないため、型にすると変換だけが増えます。検証規則があることだけを理由に型を定義しません。

**`inventory` は 2 つのエンティティを 1 つの機能パッケージに置く。** 入場枠 (`DateInventory`) と時間帯枠 (`TimeSlot`) は別の集約ですが、どちらも同じ枠在庫の減算と復元の対象になります。`domain/repository` のメソッドはエンティティ名を接頭辞に付けて (`GetDateInventory`、`ListTimeSlots`、`UpdateDateInventory`) 区別します。

**`inventory` が値オブジェクトにしたのは識別子と日付。** `ParkID`、`AttractionID`、`TimeSlotID`、`Date` は repository のメソッドに文字列を並べて渡す位置にあり、取り違えてもコンパイルが通ってしまいます。開始時刻と上限と残りは `DateInventory` と `TimeSlot` の外へ単体で出ないため、プリミティブのままにしています。

**`inventory` は枠を新しく生成しない。** `domain/model` が持つのは `RestoreDateInventory` と `RestoreTimeSlot` だけです。枠の作成は先行生成のジョブが担うため、`New` を置くと作成の経路が 2 つになります。

## 図を再作成する

```sh
# 同心円 (-p はページ番号、1 から数える)
drawio --no-sandbox -x -f png -s 2 -p 1 -o images/layers.png images/layers.drawio
```

依存グラフは `images/dependencies.svg` を HTML に包み、headless Chrome で撮ります。

```sh
printf '<!doctype html><meta charset="utf-8"><style>html,body{margin:0;background:#fff}</style>' > /tmp/dep.html
cat images/dependencies.svg >> /tmp/dep.html
"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" --headless --disable-gpu \
  --hide-scrollbars --force-device-scale-factor=2 --window-size=1420,950 \
  --default-background-color=FFFFFFFF --screenshot=images/dependencies.png /tmp/dep.html
```
