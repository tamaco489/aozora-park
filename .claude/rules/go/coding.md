# Go のコーディング規約

## パッケージの分け方

- `cmd/<service>/main.go` には設定の読み込み・依存の組み立て・サーバ (またはジョブ) の起動だけを置き、ロジックは `internal/` に配置する
- `pkg/` は設けない。外部から import される想定がないため
- `internal/` は 4 層に分ける

| ディレクトリ       | 置くもの                                                                                    |
| ------------------ | ------------------------------------------------------------------------------------------- |
| `internal/domain`  | エンティティ、ステータスとその遷移、ドメインのエラー。外部ライブラリに依存しない            |
| `internal/usecase` | ユースケース 1 つ 1 ファイル。`infra` のインタフェースを受け取り、業務の手順だけを書く      |
| `internal/infra`   | 外部サービスのラッパ (`firestore` `pubsub` `cloudtasks` `fincode` `slack`)                  |
| `internal/handler` | connect ハンドラ、Pub/Sub push・Cloud Tasks・Webhook の HTTP ハンドラ。入出力の変換に徹する |

- `gen/` は buf generate の出力。手で編集しない
- 設定は `internal/config` で 1 箇所にまとめて読む。**他のパッケージで環境変数を読まない**

## パッケージ内のファイル分割

`internal/infra/*` は次の形にする。

| ファイル       | 置くもの                                                                           |
| -------------- | ---------------------------------------------------------------------------------- |
| `doc.go`       | パッケージコメントと `package` 行だけ                                              |
| `interface.go` | 依存側のインタフェースと `var _ API = (*sdk.Client)(nil)` のような表明             |
| `client.go`    | `Client` 構造体、`Option` と `With*`、`New`、取得メソッド                          |
| `<処理>.go`    | メソッドを処理ごとに分ける。1 ファイル 1 つの関心 (`get.go` `create.go` `list.go`) |
| `errors.go`    | センチネル `var (Err...)`、専用エラー型、判定ヘルパ                                |
| `const.go`     | パッケージ定数 (コレクション名、トピック名、キュー名など)                          |

- **`fmt.Errorf` のラップとメッセージ文字列は現場に残す。** 箇所ごとに一意なメッセージであれば集約しない
- **型付き enum の定数 (`Status` `Decision` など) はその型の隣に残す。** `const.go` には移さない
- 定数と組み立て関数が一体のもの (ドキュメントパスの組み立てなど) はそのファイルに残す
- パッケージ内の型がインタフェースを満たすことの表明 (`var _ API = (*Fake)(nil)`) はその型の隣に置く
- **テストファイルは分割しない。** `<pkg>_test.go` のまま置く (テストの並びを変えるとレビューで差分が追えなくなる)

## Cloud Run の形

- service の `main.go` は `config.Load` → クライアントの組み立て → ハンドラの登録 → `http.Server` の起動の順に書き、起動の失敗は `log.Fatalf` で報告する
- 待ち受けポートは環境変数 `PORT` から読む。ハードコードしない
- SIGTERM を受けたら処理中のリクエストを待って終了する (graceful shutdown)
- job は同一バイナリのサブコマンドで切り替え、成否を終了コードで返す。常駐しない

## ハンドラの形

- connect のエラーは `connect.NewError(connect.CodeInvalidArgument, err)` のようにコードを付けて返す。ドメインのエラーからコードへの変換はインターセプタか handler 層に閉じ、usecase では素のエラーを返す
- **Pub/Sub push と Cloud Tasks は at-least-once。** すべてのハンドラを冪等にし、対象が既に終端ステータスなら何もせず 2xx を返す
- リトライしてほしい失敗だけ 5xx を返す。再実行しても直らない失敗は 2xx で ack し、ログに残す (無限リトライを避ける)
- Webhook の受信は署名検証・生保存・publish までに留め、業務ロジックを持たない
- 構造化ログに `purchaseId` `passId` `reservationId` とメッセージ ID・タスク名を必ず含める

## 外部サービスのラップ

`internal/infra/` の各パッケージは同じ形にする。

- **SDK のクライアントではなくインタフェースを受け取る。** 使うメソッドだけを列挙し、`New(api API, ...)` で組み立てる
- **エラーはセンチネルで公開する** (`ErrNotFound` など)。呼び出し側は `errors.Is` / `errors.AsType[T]` で判定でき、SDK の型に依存しない
- **コレクション名・トピック名・キュー名・宛先 URL は `New` に注入する。** パッケージ内で環境変数を読まない
- SDK が同じ意味を複数の型やコードで返す場合は、ラッパ側で 1 つのセンチネルに畳む
- Firestore のトランザクションはラッパの内側に閉じる。usecase に `*firestore.Transaction` を漏らさない

## import エイリアス

- **GCP の SDK は衝突の有無によらず `gcp` 接頭辞のエイリアスを付ける** — `gcpfirestore` `gcppubsub` `gcptasks` `gcpotel` のようにする。`internal/infra` 側と名前が衝突するため
- Firebase Admin SDK は `firebase` `firebaseauth` のようにパッケージ名を明示する
- buf が生成したコードは既定の名前 (`parkv1` `parkv1connect`) のまま使う
- **自プロジェクトのパッケージ (`internal/...`) はエイリアスを付けずパッケージ名のまま使う**

```go
import (
    gcpfirestore "cloud.google.com/go/firestore"

    "github.com/tamaco489/aozora-park/backend/internal/infra/firestore"
)
```

## 書き方

- 必要最低限の実装にする。使わない公開 API、将来用の抽象、使わない `Option` 関数を入れない
- コメントは「なぜ」が自明でないときだけ書く。句点 (。) を含めず、文の途中で改行せず 1 行で書く。長くなるなら GoDoc の箇条書き (`//   - x`) にする
- 構造体フィールドのコメントは行末に置く (`Field T // Field は ...`)
- センチネルエラーの `var (...)` は 1 件ごとに空行で区切る

## modernize の扱い

- `errors.As` ではなく `errors.AsType[T]` を使う (Go 1.26)
- ポインタが要る値は `new(x)` で作る (Go 1.26)。`Ptr` のようなヘルパは置かない
- struct には `omitempty` が効かない。`time.Time` などの struct フィールドには付けない (`omitzero` への置き換えは挙動が変わるため採らない)
- PR を出す前に `just fix-diff` と `just modernize` を通し、提案が 0 件であることを確認する

## 依存の注入

待機・時刻・乱数は外から差し替えられる形にする。
テストが実時間や実行順に依存しなくなる。

- 指数バックオフの待機は `WithSleeper`、ジッタの乱数源は `WithRandN`
- 時刻は `WithClock`。Firestore に書く `createdAt` / `updatedAt` もこれを通す

## 依存の追加

依存は必要になった時点で足す。**追加したら理由を PR に書く。**
