# Terraform のコーディング規約

## ディレクトリの分け方

- 環境は workspace ではなく `terraform/envs/{env}/` のディレクトリで分ける。当面は `dev` のみ
- 部品は `terraform/modules/{name}/` に置く。`envs/{env}/main.tf` はモジュールを呼び出すだけで、リソース定義を書かない
- provider の設定 (`provider "google" {}`、`default_labels`) と backend (GCS) は環境ディレクトリにだけ書く。モジュールは `required_providers` で要件を宣言するだけにする

## モジュールのファイル構成

| ファイル       | 置くもの                                                                        |
| -------------- | ------------------------------------------------------------------------------- |
| `versions.tf`  | `terraform { required_version, required_providers }`。provider の設定は書かない |
| `variables.tf` | 入力変数。**すべてに `description` と `type`**、任意のものだけ `default`        |
| `main.tf`      | `locals` とリソース定義                                                         |
| `outputs.tf`   | 出力値。**すべてに `description`**                                              |

- `main.tf` が肥大化するときは役割ごとに `<役割>.tf` へ分け (`api.tf` `payment_service.tf` など)、`main.tf` には `locals` と複数の役割で共有する部品だけを残す。1 ファイル 1 つの関心にし、区切り線のコメントでセクションを作らない
- Terraform 内のリソース識別子は役割名にする (`google_cloud_run_v2_service.api`、`google_pubsub_topic.purchase_created`)。同種が 1 つでも `this` にしない
- 使わない変数・出力を作らない。他モジュールが必要とする ID・名前・URL は出力し、派生形は使う側で組み立てる

## 命名と値の渡し方

- GCP プロジェクトを環境ごとに分けるため、リソース名に環境の接頭辞を付けない (`api`、`payment-service`)。プロジェクト外で一意にする必要があるもの (GCS バケットなど) だけ `-${var.project_id}` を付ける
- リソース名はモジュール内で組み立てる。環境ディレクトリは値 (`project_id` `region` `env`) を渡すだけで名前を組み立てない
- `env` は `validation` で `dev` に限る (増やすときに広げる)。`project_id` は `validation` で GCP のプロジェクト ID の形式を検査する
- **API キー・シークレット・Webhook URL・メールアドレスをファイルに書かない。** 秘匿値は Secret Manager に置き、Terraform ではシークレットの入れ物と参照だけを定義する。値の投入はユーザーが手で行う
- `terraform.tfvars` には `env` `region` のような公開してよい値だけを置く
- モジュール間の受け渡しは outputs 経由で行う。**`module` ブロックに `depends_on` を書かない** (モジュール全体の依存になり循環する)。`for_each` のキーに他モジュールの output を使わない
- ラベルは環境ディレクトリの provider `default_labels` (`project` `environment` `managed_by`) で一括付与し、モジュール内で `labels` を書かない

## 書き方

- `terraform fmt` の整形に従う (2 スペースインデント、連続する引数の `=` を揃える、ブロック間は空行 1 つ)。`just fmt-check` を通す
- 非推奨の書き方をしない。Cloud Run は `google_cloud_run_service` ではなく `google_cloud_run_v2_service` / `google_cloud_run_v2_job` を使う。IAM は `google_*_iam_policy` (全置換) ではなく `google_*_iam_member` を使う。provider の版に対する正しい書き方はドキュメントで確認する
- `google-beta` は beta 限定の引数が要るときだけ使い、その理由をコメントに残す
- バージョンは `~>` で固定する (環境側は `required_version` と provider を固定、モジュール側は下限だけで足りる)。`.terraform.lock.hcl` はコミットする

## コメント

- 「なぜ」が自明でないときだけ書く。「何をしているか」は書かない
- 日本語で書き、句点 (。) を含めない。文の途中で改行せず 1 行で書く。長くなるなら 1 行 1 論点で行を分ける
- 設計上の判断 (DLQ を付けない理由、min_instances を 0 にする理由、ingress を internal にする理由) は、**理由をリソースの直前にコメントで残す**
- `description` (変数・出力) は英語 1 文で、Terraform のドキュメント慣習に合わせる。コメントは日本語、`description` は英語と使い分ける
- 外部ツールや設計ドキュメントの所在をコメントに書かない

## 実行

- `terraform fmt` / `validate` / `plan` は実行してよい。**`terraform apply` / `destroy` はユーザーだけが実行する**
- `just lint` (tflint) と `just scan` (trivy config) は GCP に触れないためローカルで実行してよい。PR を出す前に `just fmt-check` `just validate` `just lint` `just scan` を通す
- backend を変えた直後の `init` は `-reconfigure` を使う (旧 backend に state が無いことを確認したうえで)
- Cloud Run のイメージタグは deploy ワークフローが差し替えるため、Terraform は初回作成と設定 (環境変数・シークレット参照・SA・スケール) だけを担う。`image` は `lifecycle { ignore_changes = [...] }` で無視し、CI の差し替えを drift にしない
- justfile にシェルの処理を書かない (`.claude/rules/general/justfile.md`)
