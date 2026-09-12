# proto のコーディング規約

ディレクトリは `proto/`。Protocol Buffers で API を定義し、buf で Go と TypeScript を生成する。

## 構成

### ディレクトリとパッケージ

- パッケージは `aozorapark.<サービス>.v1` の形にし、`proto/aozorapark/<サービス>/v1/` に置く。**ディレクトリとパッケージ名を一致させる**
- 末尾は必ずバージョン。`v1` を省いたパッケージを作らない
- サービス単位のパッケージと `backend/internal/<機能>` を 1 対 1 で対応させる (`aozorapark.purchase.v1` ↔ `internal/purchase`)
- ファイル名は `lower_snake_case.proto`
- 同じパッケージのファイルは同じディレクトリに置き、file option の値を揃える

### ファイルの分け方

- **1 サービス 1 ファイル**にし、その service が使う request と response を同じファイルに置く
- 複数の RPC やサービスから参照するエンティティは別ファイルに切り出す (`purchase.proto` と `purchase_service.proto`)
- 1 ファイル 1 定義 (Google の 1-1-1) は採らない。request と response は RPC 専用で他から参照されないため、分けても依存が減らない
- メッセージと enum をネストしない。後から外で参照したくなる

### ファイルの並び

`buf format` が整える順に従う。

```text
syntax → package → import (ソート済み) → file option → 定義
```

- インデントは 2 スペース、文字列は二重引用符
- `import public` と `import weak` を使わない

## 命名

### 基本規則

| 対象      | 形式                                      | 例                       |
| --------- | ----------------------------------------- | ------------------------ |
| package   | `lower_snake_case`、末尾はバージョン      | `aozorapark.purchase.v1` |
| message   | `PascalCase`                              | `Purchase`               |
| field     | `lower_snake_case`                        | `purchase_id`            |
| enum      | `PascalCase`                              | `PurchaseStatus`         |
| enum の値 | `UPPER_SNAKE_CASE`、enum 名を接頭辞にする | `PURCHASE_STATUS_PAID`   |
| service   | `PascalCase`、`Service` で終える          | `PurchaseService`        |
| rpc       | `PascalCase`                              | `CreatePurchase`         |

- **enum のゼロ値は `_UNSPECIFIED` で終える。** 未設定と既定値が区別できなくなるため
- 略語は 1 語として扱う (`GetDnsRequest` であって `GetDNSRequest` ではない)
- 下線の後には必ず英字を置く (`XYZ2` であって `XYZ_2` ではない)。言語ごとの変換で名前が衝突する
- `repeated` のフィールドは複数形にする
- 型名がそのまま使える場合はフィールド名を型名に合わせる (`Purchase purchase = 1;`)
- 言語の予約語を型名とパッケージ名に使わない。`internal` は Go の生成コードが import できなくなる

### RPC と入出力の型

- **RPC ごとに専用の request と response を作る。** 他の RPC と共有しない
- 名前は `<RPC 名>Request` と `<RPC 名>Response`
- **中身が空でも `google.protobuf.Empty` を使わない。** 空のメッセージを自分で定義する。後からフィールドを足すときに破壊的変更にならない
- **ストリーミング RPC を使わない。** 非同期の処理は Pub/Sub と Cloud Tasks で行う

## 型の選び方

| 表したいもの | 使う型                      |
| ------------ | --------------------------- |
| 時刻         | `google.protobuf.Timestamp` |
| 期間         | `google.protobuf.Duration`  |
| 金額         | `int64` で最小単位 (円)     |
| ID           | `string`                    |
| 状態・分類   | `enum`                      |

- **金額に `double` と `float` を使わない。** 丸め誤差が出る
- **2 値でも、将来 3 値になりうるものに `bool` を使わない。** `enum` にする
- フィールド番号の 1 から 15 は 1 バイトで符号化される。頻出するフィールドに割り当てる
- `optional` は、未設定と既定値を区別したいときだけ付ける。区別が要らないなら付けない

## 変更のしかた

**クライアントとサーバは同時に入れ替わらない。** 片方が古いまま動く前提で変更する。

- **フィールド番号を再利用しない。** 既存のデータが別の意味で読まれる
- フィールドを削除したら `reserved` に番号と名前を残す (`reserved 3; reserved "old_name";`)
- enum の値を削除したときも同じく `reserved` に残す
- **フィールドの型を変えない。** 番号の再利用と同じ問題が起きる
- `repeated` と単数の相互変更をしない
- 破壊的変更は `buf breaking` で検出する。CI への導入はサブ Issue #7
- どうしても壊す必要があるなら `v2` のパッケージを新しく作り、`v1` は残す

## コメント

- `//` を使う。`/* */` は使わない
- 型・フィールド・RPC の**上**に書く。行末に書かない
- 日本語で書き、句点 (。) を含めない。1 行で書く
- **生成コードのドキュメントとして外に出る。** 読み手が実装を見られない前提で書く
- 単位と制約はコメントに書く (`// amount は税込みの総額、単位は円`)

## エラー

**エラーの表現を proto に持ち込まない。** 成否を表すフィールドや、独自のエラーメッセージを response に入れない。

connect のエラーコードとアプリケーションのコードは `platform/serving/apperr` とインターセプタが担当する (`.claude/rules/go/coding.md`)。

## 生成と実行

- `just lint` `just fmt` `just fmt-check` `just generate` を `proto/` で実行する。PR を出す前に `just lint` と `just fmt-check` を通す
- プラグインは BSR のリモート版を使う。ローカルに `protoc-gen-*` を入れない
- **生成物は `backend/gen/` と `frontend/src/gen/` にコミットし、手で編集しない**
- **`.proto` を削除しても生成ファイルは残る。** リネーム・削除をしたら出力先の不要なファイルを手で消す

> [!NOTE]
> 入力値の検証 (protovalidate) はまだ導入していない。最初の業務 API を作る時点で、
> どこまでを proto の制約として書き、どこからを Go の `handler` と `domain/model` に
> 置くかを決めてここに書き足す。
