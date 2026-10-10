# backend のパッケージ構成

[English](./README.md) | [日本語](./README.ja.md)

[ドキュメント一覧](../../README.ja.md)に戻る。

`backend/` の内部構造をまとめます。規約は `.claude/rules/go/coding.md` が持ちます。ここには現状と、その形を選んだ理由を置きます。

## 層と依存の向き

外側をビジネス機能で分け、機能の内側を層で分けます。依存は内側に向かうものだけを許します。

![層と依存の向き](./images/layers.png)

| リング                     | パッケージ                                                                                                        |
| -------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| Enterprise Business Rules  | `<機能>/domain/model`、`platform/serving/apperr`                                                                  |
| Application Business Rules | `<機能>/usecase`、`<機能>/usecase/port`、`<機能>/domain/repository`                                               |
| Interface Adapters         | `<機能>/handler`、`<機能>/infrastructure/firestore`、`<機能>/infrastructure/pubsub`、`internal/<機能>` (組み立て) |
| Frameworks & Drivers       | `cmd/api`、`cmd/job`、`internal/platform/**`、`gen/**`                                                            |

機能パッケージは `park`、`inventory`、`prioritypass` の 3 つで、どれも同じ形をとります。
`usecase/port` と `infrastructure/pubsub` を持つのは `prioritypass` だけです。

`cmd/job` は 1 つのバイナリをサブコマンドで切り替えます。今あるのは枠の先行生成 (`generate`) だけです。

優先パスの割当を受ける `cmd/priority-pass-issuer` と、Pub/Sub push のエンベロープを解く `platform/serving/pubsubpush` はまだ置いていません。
置き場所だけを `.claude/rules/go/coding.md` で決めています。

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

**`prioritypass` だけが `usecase/port` を持つ。** 永続化は `domain/repository` に置きますが、publish は永続化ではありません。使う側が必要とするメソッドだけを `usecase/port` に並べ、実装は `infrastructure/pubsub` が持ちます。`park` と `inventory` は Firestore しか触らないため、この層がありません。

**publish の失敗で申込を巻き戻さない。** 作成済みの申込を呼び出し元が作り直すと二重の申込になるため、送信の失敗はエラーとして返さずログに残します。送れたことは `publishedAt` に記録し、残らなかったものを後から送り直せるようにしています。

**publish は呼び出し元の ctx から切り離し、待つ上限を置く。** 宛先に届かない間 SDK がリトライを続けるため、上限を置かないと申込を返すのが遅れます。実測では、上限を置かない場合に 60.3 秒かかった RPC が、`context.WithoutCancel` と 3 秒の `context.WithTimeout` で 3.02 秒になりました。

**機能パッケージ同士を import せず、相手のドキュメントは自分の構造体で読む。** `inventory` は `park` が書いたマスタを、`prioritypass` は `inventory` が書いた時間帯枠を読みます。どちらも相手の `domain/model` を使わず、必要な項目だけの構造体を自分側に置きます。`DataTo` は構造体に無い項目を捨てるため、相手が項目を増やしても壊れません。

引き換えに、コレクションのパスとフィールド名が 2 か所に並びます。片方だけを書き換えても、実装とテストのヘルパが同じ定数を参照していると気づけないため、組み立てたパスをリテラルと突き合わせるテストを置いています (`TestRepositoryTimeSlotDocPath`)。3 つ目の利用者が現れたら、この形のままでよいかを見直します。

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
