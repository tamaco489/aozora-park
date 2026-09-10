# Go のコーディング規約

## 構成

### アーキテクチャ

**Package by Feature を採る。** `internal/` を層で分けず、業務機能で分ける。層で分けるのは機能パッケージの中だけにする。

Package by Layer (`internal/{domain,usecase,infra,handler}`) は採らない。1 つの機能を追うのに 4 つのディレクトリを行き来することになり、変更の単位とディレクトリの単位が一致しないため。このプロジェクトはマイルストーンが機能単位で進むので、1 機能が 1 ディレクトリに収まる形にする。

外側の分割が Package by Feature、内側の分割がレイヤードという二段構えになる。守るのは**層の数ではなく依存の向き**で、それは「依存の向き」で定める。

### ディレクトリ構成

```text
backend/
├── cmd/                              # バイナリ。4 段の組み立てだけを書く
│   ├── api/
│   ├── webhook-receiver/
│   ├── payment-service/
│   ├── notification-service/
│   ├── allocation-service/
│   ├── group-reservation-service/
│   └── job/                          # サブコマンドで expire / summary / generate / reconcile を切替
├── internal/
│   ├── purchase/                     # 業務機能 1 つ。他の機能パッケージを import しない
│   │   ├── purchase.go               # 組み立て関数だけを置く
│   │   ├── domain/
│   │   │   ├── model/                # エンティティ・値オブジェクト・ステータス遷移
│   │   │   │   ├── purchase.go
│   │   │   │   ├── status.go
│   │   │   │   └── errors.go         # センチネルエラーとコード
│   │   │   └── repository/           # 永続化のインタフェース
│   │   │       ├── reader.go         # 参照のみ
│   │   │       └── writer.go         # 更新のみ
│   │   ├── usecase/                  # 1 ユースケース 1 ファイル
│   │   │   ├── port/                 # 永続化以外の外部サービス
│   │   │   │   ├── publisher.go
│   │   │   │   └── payment.go
│   │   │   ├── create.go
│   │   │   └── issue_ticket.go
│   │   ├── infrastructure/           # domain/repository の実装
│   │   │   └── firestore/
│   │   └── handler/                  # 入口
│   │       ├── connect.go
│   │       └── subscriber.go         # Pub/Sub push
│   ├── park/                         # 以下、機能パッケージは同じ形にする
│   ├── prioritypass/
│   ├── groupreservation/
│   ├── payment/
│   ├── notification/
│   ├── inventory/                    # 複数の機能から使う枠在庫の減算・復元
│   └── platform/                     # 機能に依存しない共通基盤
│       ├── config/                   # 環境変数の読み込み。ここ以外では読まない
│       ├── client/                   # プロセスの外へ出ていくもの
│       │   ├── firestore/
│       │   ├── pubsub/
│       │   ├── cloudtasks/
│       │   ├── auth/
│       │   ├── fincode/
│       │   └── slack/
│       ├── serving/                  # リクエストを受ける側の下回り
│       │   ├── httpx/                # サーバ起動・graceful shutdown・App
│       │   ├── interceptor/          # 認証・ログ・エラー変換
│       │   └── apperr/               # エラーの型・Kind・connect.Code への変換
│       └── observability/            # 起きたことを外に出すもの
│           ├── logging/
│           └── telemetry/
└── gen/                              # buf generate の出力。手で編集しない
    └── aozorapark/<サービス>/v1/
```

### パッケージの分け方

- `cmd/<service>/main.go` には設定の読み込み・依存の組み立て・サーバ (またはジョブ) の起動だけを置き、ロジックは `internal/` に配置する
- `pkg/` は設けない。外部から import される想定がないため
- **`internal/` は層ではなく業務機能で分ける。** 機能パッケージと、機能に依存しない `platform` に分かれる

| ディレクトリ         | 置くもの                                                        |
| -------------------- | --------------------------------------------------------------- |
| `internal/<機能>`    | 業務機能 1 つ (`park` `purchase` `prioritypass` `payment` など) |
| `internal/inventory` | 複数の機能から使う枠在庫の減算・復元                            |
| `internal/platform`  | 機能に依存しない共通基盤                                        |
| `gen/`               | buf generate の出力。手で編集しない                             |

- **機能パッケージ同士は import しない。** 共有が必要になったら `inventory` のように独立したパッケージへ切り出し、利用側はインタフェースで受け取る
- `platform` の直下は `config` と 3 つのグループだけにする

| グループ        | 入れるもの                                                 | 判断基準                     |
| --------------- | ---------------------------------------------------------- | ---------------------------- |
| `client`        | `firestore` `pubsub` `cloudtasks` `auth` `fincode` `slack` | プロセスの外へ出ていくもの   |
| `serving`       | `httpx` `interceptor` `apperr`                             | リクエストを受ける側の下回り |
| `observability` | `logging` `telemetry`                                      | 起きたことを外に出すもの     |

- 設定は `internal/platform/config` で 1 箇所にまとめて読む。**他のパッケージで環境変数を読まない**

### 機能パッケージの中

機能で切ったうえで、その中を 4 層に切る。

| ディレクトリ        | 置くもの                                                               |
| ------------------- | ---------------------------------------------------------------------- |
| `domain/model`      | エンティティ、値オブジェクト、ステータス遷移、センチネルエラーとコード |
| `domain/repository` | 永続化のインタフェース。`reader.go` と `writer.go` に分ける            |
| `usecase`           | 業務の手順。1 ユースケース 1 ファイル                                  |
| `usecase/port`      | 永続化以外の外部サービスのインタフェース (publish・決済・通知)         |
| `infrastructure`    | `domain/repository` の実装 (`firestore/`)                              |
| `handler`           | 入口。connect ハンドラ、Pub/Sub push、Cloud Tasks、Webhook             |

- 機能パッケージのルートには組み立て関数だけを置く (「依存の組み立て」を参照)
- `handler` は入口ごとにファイルを分け (`connect.go` `subscriber.go`)、どれも同じユースケースを呼ぶ
- **テストファイルは分割しない。** 対象と同じディレクトリに `<pkg>_test.go` として置く

### インタフェースの置き場所

- **永続化は `domain/repository`。** Reader と Writer に分け、参照しかしないユースケースに更新手段を渡さない
- **それ以外の外部サービスは `usecase/port`。** 使う側が必要とするメソッドだけを並べる
- 実装は 1 つの構造体が Reader と Writer の両方を満たしてよい。分けるのは受け取る側の都合で、実装を分けるためではない

### usecase は素通しでも置く

`handler` から `infrastructure` を直接呼ばない。取得して返すだけの処理でも `usecase` を通す。

素通しは長く続かない。冪等性の判定、権限の確認、複数の集約への操作は後から必ず入る。そのときに層を挿し込むより、最初から通しておくほうが安い。`handler` がドメインの型を直接触らずに済む効果もある。

### 依存の向き

許される向きは 4 つだけ。

```text
handler        → usecase → domain
infrastructure → domain
すべての層     → platform
cmd            → すべて
```

禁止するもの。

- `domain` が `usecase` や `infrastructure` を import する
- `infrastructure` が `usecase` を import する (インタフェースは `domain/repository` にある)
- `usecase` が connect の型や `gen/` を import する (入出力の変換は `handler` の仕事)
- 機能パッケージ同士の import
- `platform` が `internal/<機能>` を import する

`cmd` だけが全層を import してよい。例外ではなく、最も外側だから許される。

**この向きは golangci-lint の depguard で検査する。** 規約だけ置いて検査しない期間を作らない。

### import エイリアス

- **サブパッケージ名は機能をまたいで衝突するため `<機能><層>` でエイリアスする** — `parkusecase` `purchasehandler` `purchasemodel`
- **GCP の SDK は衝突の有無によらず `gcp` 接頭辞を付ける** — `gcpfirestore` `gcppubsub` `gcptasks` `gcpotel`
- Firebase Admin SDK は `firebase` `firebaseauth` のようにパッケージ名を明示する
- buf が生成したコードは既定の名前 (`parkv1` `parkv1connect`) のまま使う

```go
import (
    gcpfirestore "cloud.google.com/go/firestore"

    purchasemodel "github.com/tamaco489/aozora-park/backend/internal/purchase/domain/model"
    purchaseusecase "github.com/tamaco489/aozora-park/backend/internal/purchase/usecase"
)
```

## ドメインの設計

### 不変条件を守る

**`domain/model` のフィールドは公開しない。** 状態は振る舞いを表すメソッドで変える。

```go
type Purchase struct {
    id     PurchaseID
    status Status
    items  []Item
}

func (p *Purchase) Confirm(now time.Time) error { ... }
func (p *Purchase) Cancel(reason string) error  { ... }
```

- フィールドを公開すると `p.Status = StatusPaid` が書けてしまい、ステータス遷移の規則が意味を失う
- **すべてのフィールドにゲッターを作らない。** 外に出す必要があるものだけ公開する。ゲッターを増やすと判断が `handler` に漏れる
- ゲッターに `Get` 接頭辞を付けない (`p.Status()` であって `p.GetStatus()` ではない)
- 不変条件を持たない入れ物 (設定・DTO・リクエスト) はフィールドを公開してよい。すべてを非公開にするわけではない
- 生成した時点で妥当な状態になるようにする。`New` が検証し、不正な入力ではインスタンスを返さない

### 値オブジェクト

**次のどれかに当てはまるものだけ型にする。**

- 検証規則がある (`ISBN` `Email`)
- 型の取り違えを防ぎたい (`PurchaseID` と `PassID` を混ぜない)
- ドメイン固有の振る舞いがある (`Money.Add`)
- 正規化が必要 (前後の空白、大文字小文字)

当てはまらないものはプリミティブのままにする。すべてのフィールドを型にすると変換だけが増え、`user.Name().Value()` のような呼び出しが並ぶ。**過剰な値オブジェクトは、プリミティブの多用と同じくらい避ける。**

> [!NOTE]
> ここは実装が始まってから見直す。実際に作った型を見て、基準が厳しすぎるか緩すぎるかを判断し、このプロジェクトの実物に合わせて書き直す。

### 定数の定義

**`iota` を使わない。値を 1 つずつ明示する。**

`iota` は並びに意味を持たせるため、途中の 1 件を削除したり順序を入れ替えたりすると、後続の値が黙ってずれる。ステータスやエラーの分類は Firestore に保存され、ログにも出て、外部にも送られるため、値がずれると既存データの意味が変わる。コンパイルは通り、テストも落ちないまま壊れる。

```go
// 良い例。値が独立しているので、削除しても並び替えても他に影響しない
type Status string

const (
    StatusPending Status = "pending"
    StatusPaid    Status = "paid"
    StatusExpired Status = "expired"
)
```

```go
// 悪い例。StatusPaid を削除すると StatusExpired の値が 2 から 1 に変わる
type Status int

const (
    StatusPending Status = iota + 1
    StatusPaid
    StatusExpired
)
```

- 永続化・送信・ログに出る値は文字列で定義する。数値の連番にしない
- 数値が要る場合 (優先度・上限など) も `= 1` `= 2` のように値を書く
- 例外は、外部に出ず順序そのものに意味がある内部的なビット列だけ。使うときは理由をコメントに残す

### バリデーション

3 層で分担する。同じ検証が 2 層に現れてよい。入口は connect だけでなく Pub/Sub や Cloud Tasks もあるため。

| 層             | 検証するもの                                   | 例                                    |
| -------------- | ---------------------------------------------- | ------------------------------------- |
| `handler`      | 形式と型                                       | 必須フィールド、文字列長、日付の形式  |
| `usecase`      | ユースケース固有の整合性。外部の状態を見る判断 | 重複、権限、対象が操作可能な状態か    |
| `domain/model` | 不変条件。その型が単体で常に満たすべき規則     | 数量が 1 以上、ステータス遷移の妥当性 |

- `domain/model` の検証は省略しない。他の層を通さない経路 (ジョブ・再実行) があるため
- 形式の検証を `usecase` に書かない。逆に業務規則を `handler` に書かない

### トランザクションの境界

**1 トランザクションで 1 集約しか更新しない。**

- またぎたくなったら、まず集約の境界が間違っていないかを疑う
- それでも複数を更新する必要があるなら、Firestore のトランザクションで束ねずに Pub/Sub で分ける。購入の確定と Slack 通知のように、片方が失敗しても再実行で追いつく形にする
- 即座の整合性が要る組み合わせは、同じ集約に入れることを検討する
- トランザクションは `infrastructure` の内側に閉じる。`usecase` に `*firestore.Transaction` を渡さない
- イベント名は過去形にする (`purchase.created` `payment.succeeded`)。発行済みのイベントの意味を後から変えない

## エラー

### エラーの扱い

エラーに関わる場所は 3 つあるが、責務は重ならない。

| 場所                            | 持つもの                                                                     |
| ------------------------------- | ---------------------------------------------------------------------------- |
| `platform/serving/apperr`       | エラーの型、分類 (`Kind`)、`Kind` から `connect.Code` への変換、リトライ可否 |
| `<機能>/domain/model/errors.go` | その機能のコード (`PURCHASE_SOLD_OUT`) と、属する `Kind`                     |
| `platform/serving/interceptor`  | 実際に変換してレスポンスとログに載せる唯一の場所                             |

```go
// internal/purchase/domain/model/errors.go
var ErrSoldOut = apperr.New(apperr.KindConflict, "PURCHASE_SOLD_OUT", "在庫が不足している")
```

- **`Kind` を増やせるのは `apperr` だけ、コード文字列を増やせるのは機能側だけ**
- コードの接頭辞は機能名に固定する。重複は全コードを集めるテストで検出する
- **`infrastructure` は SDK のエラーを domain のセンチネルに翻訳する。** 見つからないは `model.ErrXxxNotFound` に畳み、それ以外は `%w` でラップして伝播する
- `infrastructure` が HTTP や connect のコードを知ることはない
- `usecase` はセンチネルをそのまま返し、`handler` は素通しする
- Pub/Sub push と Cloud Tasks のハンドラは `apperr.Retryable` を見て、5xx で返すか 2xx で ack するかを決める

## 外部との境界

### 外部サービスのラップ

`internal/platform/client/` の各パッケージは同じ形にする。

- **SDK のクライアントではなくインタフェースを受け取る。** 使うメソッドだけを列挙し、`New(api API, ...)` で組み立てる
- **業務判断を持たない。** コレクション名・トピック名・キュー名・宛先 URL は `New` に注入する。パッケージ内で環境変数を読まない
- SDK が同じ意味を複数の型やコードで返す場合は、ラッパ側で 1 つのセンチネルに畳む
- Firestore のトランザクションは `infrastructure` の内側に閉じる。`usecase` に `*firestore.Transaction` を漏らさない

### 外部 API の翻訳

**外部サービスの型を `usecase` より内側に持ち込まない。** fincode のレスポンス構造体、Slack のペイロード、Pub/Sub のメッセージ型は `platform/client` と機能パッケージの境界で自分たちの型に翻訳する。

- 翻訳の場所は `infrastructure` か、`usecase/port` の実装側
- 外部の項目名や省略形をそのままドメインの語彙にしない
- 翻訳が必要ないほど素直な相手 (自分たちで定義した Pub/Sub のメッセージなど) にまで層を挟まない

> [!NOTE]
> ここも実装が始まってから見直す。fincode の実際のレスポンスを見て、翻訳の粒度と置き場所をこのプロジェクトの形に合わせる。

## 組み立てと実行

### 依存の組み立て

**DI に外部ライブラリを使わない。** コンストラクタで手書きする。

- 機能パッケージのルートに組み立て関数を置く (`purchase.NewConnectHandler(...)`)。`usecase` と `infrastructure` と `handler` の結線はここに閉じる
- `cmd/*/main.go` は 4 段に固定する。設定を読む → クライアントを作る → 機能の組み立て関数を呼ぶ → サーバを起動する
- クライアントの終了処理は `platform/serving` の `App` に登録し、登録の逆順に閉じる
- 業務上の時刻 (`createdAt`、失効の基準時刻) と乱数は `WithClock` `WithRandN` で差し替えられるようにし、組み立て関数の引数か `Option` で渡す
- **待機のための `Sleeper` は作らない。** バックオフやポーリングは `time.Sleep` をそのまま書き、テスト側を `testing/synctest` で囲む (`.claude/rules/go/testing.md`)
- **`main` が 4 段の形を保てなくなったら、まず組み立てを機能パッケージ側へ押し戻す。** 実行時解決のコンテナ (fx・dig) は採らない

### Cloud Run の形

- service の `main.go` は上記 4 段の順に書き、起動の失敗は `log.Fatalf` で報告する
- 待ち受けポートは環境変数 `PORT` から読む。ハードコードしない
- SIGTERM を受けたら処理中のリクエストを待って終了する (graceful shutdown)
- job は同一バイナリのサブコマンドで切り替え、成否を終了コードで返す。常駐しない

### ハンドラの形

- connect のエラーはインターセプタが `apperr` から変換する。`handler` で `connect.NewError` を組み立てない
- **Pub/Sub push と Cloud Tasks は at-least-once。** すべてのハンドラを冪等にし、対象が既に終端ステータスなら何もせず 2xx を返す
- リトライしてほしい失敗だけ 5xx を返す。再実行しても直らない失敗は 2xx で ack し、ログに残す (無限リトライを避ける)
- Webhook の受信は署名検証・生保存・publish までに留め、業務ロジックを持たない
- 構造化ログに `purchaseId` `passId` `reservationId` とメッセージ ID・タスク名を必ず含める

## 書き方

### 実装とコメント

- 必要最低限の実装にする。使わない公開 API、将来用の抽象、使わない `Option` 関数を入れない
- **実装が 1 つしかないインタフェースを先回りで作らない。** 差し替えるか、テストでフェイクに置き換えるものだけ定義する
- コメントは「なぜ」が自明でないときだけ書く。句点 (。) を含めず、文の途中で改行せず 1 行で書く。長くなるなら GoDoc の箇条書き (`//   - x`) にする
- 構造体フィールドのコメントは行末に置く (`Field T // Field は ...`)
- センチネルエラーの `var (...)` は 1 件ごとに空行で区切る

### modernize の扱い

- `errors.As` ではなく `errors.AsType[T]` を使う (Go 1.26)
- ポインタが要る値は `new(x)` で作る (Go 1.26)。`Ptr` のようなヘルパは置かない
- struct には `omitempty` が効かない。`time.Time` などの struct フィールドには付けない (`omitzero` への置き換えは挙動が変わるため採らない)
- PR を出す前に `just fix-diff` と `just modernize` を通し、提案が 0 件であることを確認する

### 依存の追加

依存は必要になった時点で足す。**追加したら理由を PR に書く。**
