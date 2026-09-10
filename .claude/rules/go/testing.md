# Go のテスト規約

## 方針

### 実サービスを呼ばない

テストはフェイクかエミュレータで行う。**GCP・fincode・Slack の実 API を呼ばない。**

### 層ごとの検証

| 層               | 何で検証するか                | 何を確かめるか                                     |
| ---------------- | ----------------------------- | -------------------------------------------------- |
| `domain/model`   | 実物だけ                      | 不変条件、ステータス遷移、値オブジェクトの検証規則 |
| `usecase`        | フェイク                      | 業務の手順、分岐、エラーの種類                     |
| `infrastructure` | エミュレータ (testcontainers) | 保存と取得、クエリ、トランザクション、エラーの翻訳 |
| `handler`        | フェイク                      | 入出力の変換、冪等性、リトライさせるかの判断       |

- **ドメインの型はモックしない。** エンティティと値オブジェクトは実物を使う
- 差し替えるのは `domain/repository` と `usecase/port` のインタフェース、それと `platform/client` だけ
- `domain/model` と `usecase` のテストは Docker に依存させない。フェイクだけで完結させる

### カバレッジに目標値を決めない

低い数字は「広い範囲が自動テストで触られていない」ことを示すので使う。高い数字は品質を保証しない。

**見るのは「覆われている割合」ではなく「覆われていない場所」。** ステータス遷移、在庫の減算、決済の分岐が抜けていないかを確かめる。

## 書き方

### テーブル駆動テストの書き方

- ケースは `map[string]struct{...}` で持ち、`for name, tt := range tests { t.Run(name, ...) }` で回す
- キーは `正常系_xxxの場合_xxxになること` / `異常系_...` / `境界値_...` の 3 部構成にする
- 単一のシナリオを一続きで確かめるテストは無理にテーブルにしない
- ステータス遷移は全パターンを表にし、遷移図と突き合わせる

補足。ケース名にスペースとスラッシュを入れない。スペースは `_` に置換され、スラッシュはサブテストの階層区切りとして解釈されるため、`-run` での指定が壊れる。区切りは `_` を使う。

`tt := tt` は書かない。Go 1.22 以降、ループ変数は反復ごとに新しく作られる。

### 失敗メッセージ

**`関数名(入力) = got, want want` の順で書く。** got が先、want が後。

```go
t.Errorf("Purchase.Confirm(%v) = %v, want %v", in, got, want)
t.Errorf("CreatePurchase() の差分 (-want +got):\n%s", cmp.Diff(want, got))
```

- 関数名と入力を必ず含める。テーブルの添字だけでは何が壊れたか分からない
- `cmp.Diff` は `cmp.Diff(want, got)` の順に呼び、メッセージに `-want +got` と書く
- 失敗を読むのが自分とは限らない前提で書く

### アサーションと比較

- **アサーションライブラリを入れない** (testify など)。Go の `if` と `t.Errorf` で書く
- 構造体の比較は `reflect.DeepEqual` ではなく `google/go-cmp` の `cmp.Diff` を使う。未公開フィールドの変更に振り回されない
- `cmp` はテスト専用。本番コードで使わない (未公開フィールドがあると panic する)
- フィールドを 1 つずつ比べず、構造体ごと比較して差分を出す
- **エラーは文字列で比較しない。** `errors.Is` でセンチネルを、`errors.AsType[T]` で型を判定する

### テストヘルパ

- ヘルパは `*testing.T` ではなく `testing.TB` を受け取る。Test と Benchmark と Fuzz から使い回せる
- 冒頭で `tb.Helper()` を呼ぶ。失敗の行がヘルパの中ではなく呼び出し側に出る
- **後始末は `defer` ではなく `t.Cleanup` に登録する。** ヘルパ内の `defer` はヘルパが return した時点で走ってしまう。並列サブテストがある場合も `t.Cleanup` でないと早すぎる
- セットアップの失敗は `t.Fatal` でよい。検証の失敗は `t.Error` にして、1 回の実行で全部出す
- **`t.Fatal` と `t.FailNow` をテスト以外の goroutine から呼ばない。** `runtime.Goexit` を呼ぶだけでテスト本体に失敗が伝わらない。goroutine の中では `t.Error` を使う
- **`t.Cleanup` は panic では走らない。** コンテナや外部リソースの解放をこれだけに頼らない

### Fuzz

**外部から任意の入力を受ける境界にだけ書く。** Webhook のペイロードのパース、署名検証、Pub/Sub メッセージのデコードが対象になる。

- `func FuzzXxx(f *testing.F)` に `f.Add` でシードを与える
- 失敗した入力は `testdata/fuzz/FuzzXxx` に自動で書き出される。**これはコミットする。** 以後は通常の `go test` でも回帰テストとして走る
- fuzz target は決定的にする。永続状態やグローバルに依存させない

### テストの主張が空振りしていないか確かめる

テストを書いたら、**検証対象を意図的に壊して落ちることを確認する。**

通ることだけを確認したテストは、何も検証していない可能性がある。

## テストダブル

### フェイクの置き場所

**テストファイルの中に定義する。** インタフェースは小さいので、10 行程度の構造体で足りる。モック生成ツールは使わない。

```go
type fakePurchaseReader struct {
    purchases map[model.PurchaseID]*model.Purchase
}

var _ repository.Reader = (*fakePurchaseReader)(nil)
```

- `var _ Interface = (*fake)(nil)` を書いておくと、インタフェースの変更にフェイクが追従していないことをコンパイル時に検出できる
- 別パッケージに出すのは、`platform/client` のフェイクを複数の機能から使う場合だけにする
- 呼び出し回数の検証は最小限にする。実装の詳細に依存するとリファクタリングで壊れる

### HTTP を挟むテスト

- fincode と Slack の呼び出しは `httptest` のサーバに向ける
- サーバへ投げるクライアントは自前の `Transport` を組まず `server.Client()` を使う
- `httptest.ResponseRecorder` は 1 レスポンス分しか記録しない。使い回さない
- Go 1.27 の `httptest.NewTestServer` は終了時の片付けを自分で登録するので `Close` を書かなくてよい

### 外部 API の記録

fincode の応答をフィクスチャに置く場合は `backend/testdata/fincode/` にまとめる。

**実レスポンスでない記録には `note` フィールドでその旨を明記する。**
手書きや合成のフィクスチャを実物と誤認すると、通っているテストが何も保証しなくなる。

## 揺らぎを持ち込まない

### 実時間を待たない

**時間に依存するテストは `testing/synctest` で書く。** `time.Sleep` で実際に待たない。

```go
synctest.Test(t, func(t *testing.T) {
    // bubble の中では時計が仮想。24 時間の待機も一瞬で終わる
    synctest.Sleep(24 * time.Hour)
    synctest.Wait() // 他の goroutine が落ち着くまで待つ
    // ここで状態を検査する
})
```

- 指数バックオフ、TTL の失効、Cloud Tasks の遅延、ポーリングはこれで書く
- **durably block しないものがある。** mutex、ネットワーク I/O、システムコールは対象外。実 I/O を含む処理は bubble の中に入れず、フェイクに置き換える
- bubble の外から bubble 内のチャネルやタイマーを触ると panic する
- 乱数は `WithRandN` で差し替え、下限・中央・上限を決定的に検証する
- **業務上の時刻は `Clock` の差し替えを続ける。** Firestore に書く `createdAt` や、失効判定の基準時刻のように、値そのものを検証したいものは synctest の仮想時計では扱えない

### 冪等性を必ず検証する

Pub/Sub と Cloud Tasks は at-least-once。
**worker のテストは同じメッセージを 2 回渡し、2 回目で状態が変わらないことを確かめる。**

在庫の減算は、同時実行で残数がマイナスにならないこと・終端ステータスからの再遷移が起きないことを検証する。

### goroutine を残さない

Pub/Sub の購読、リトライ、graceful shutdown は goroutine を起こす。終了時に残っていないことを検査する。

`go.uber.org/goleak` の `goleak.VerifyTestMain(m)` を `TestMain` に置く。個別のテストでの `VerifyNone` は並列テストで誤検知するため使わない。

## 外部依存のあるテスト

### エミュレータを使うテスト

**`infrastructure` のテストは testcontainers-go でエミュレータを起動する。** `github.com/testcontainers/testcontainers-go/modules/gcloud` に Firestore と Pub/Sub のコンテナがある。

- Docker がない環境のために環境変数でゲートし、未設定なら `t.Skip("run with DOCKER_TESTS=1")` する。**CI では必ず設定して走らせる**
- ビルドタグ (`//go:build integration`) では分けない。タグ付きファイルは gopls と golangci-lint の対象から外れ、気づかないうちに腐る
- コンテナはパッケージ単位で `TestMain` から 1 つ起動して使い回す。テストごとの起動は遅すぎる
- **`go test ./...` はパッケージを並列に実行する。** エミュレータを使うパッケージが増えたら `-p 1` を検討する
- テスト間の分離は、コレクション名を分けるか `t.Cleanup` で消して行う。`t.Cleanup` は panic では走らないため、次の実行が前のデータに影響されない書き方にする

> [!NOTE]
> 導入はサブ Issue #5 以降。Firestore を触るコードができた時点で、起動の作法とクリーンアップの形をここに書き足す。

`docker compose` のエミュレータはテスト用ではない。アプリをローカルで動かして手で叩くためのもので、用途が違う。

## 実行と資材

### テストの実行

```sh
go test ./...                                  # 日常
go test -race -shuffle=on -count=1 ./...       # CI
```

- `-shuffle=on` はテスト間の順序依存を検出する。失敗時は出力された seed で再現できる
- `-count=1` はテスト結果のキャッシュを無効にする。エミュレータの状態はキャッシュ判定に入らないため、CI では必ず付ける
- `-race` はメモリを 5 倍から 10 倍、実行時間を 2 倍から 20 倍にする。遅ければ別ジョブに分ける
- カバレッジを取るときは `-covermode=atomic` にする。`-race` と併用できる書式はこれだけ

### テストデータの置き場所

| 置き場所                       | 内容                                              |
| ------------------------------ | ------------------------------------------------- |
| 各パッケージの `testdata/`     | そのパッケージ内で完結するデータ                  |
| `backend/testdata/<サービス>/` | 外部 API の記録済みレスポンス。複数の機能から使う |

テストがリポジトリ内のファイルを書き換えない。出力先は `t.TempDir()` を使う。例外は fuzz が `testdata/fuzz` に書く場合だけ。
