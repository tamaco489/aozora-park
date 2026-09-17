# インフラのコーディング規約

ディレクトリは `infra/`。Terraform で Google Cloud のリソースを定義する。

## ディレクトリの分け方

- 環境は workspace ではなく `infra/envs/{env}/` のディレクトリで分ける。GCP に置く環境は `stg` と `prd` で、`dev` はローカル専用のため持たない
- 環境の識別子は `dev` `stg` `prd` に固定する。`prod` `staging` `development` を使わない
- `prd` は構成の定義だけを持ち、apply しない。GCP のプロジェクトと state バケットも作成しないため、`backend.tf` を置かず、検査は `init` の上で `validate` まで行う
- 部品は `infra/modules/{name}/` に置く。`envs/{env}/main.tf` はモジュールを呼び出すだけで、リソース定義を書かない
- provider の設定 (`provider "google" {}`、`default_labels`) と backend (GCS) は環境ディレクトリにだけ書く。モジュールは `required_providers` で要件を宣言するだけにする

## Terraform で管理しないもの

- GCP の組織・フォルダ・プロジェクト、課金アカウントの紐付け、state バケット、予算アラートは手作業で作成する。手順は設計ドキュメントに置く
- Terraform はプロジェクトの中のリソースだけを扱う

## 環境ディレクトリのファイル構成

| ファイル              | 置くもの                                                                            |
| --------------------- | ----------------------------------------------------------------------------------- |
| `versions.tf`         | `terraform { required_version, required_providers }`                                |
| `backend.tf`          | GCS の backend。`stg` だけに置く                                                    |
| `providers.tf`        | `provider "google" {}` と `default_labels`                                          |
| `variables.tf`        | `project_id` `region` `env` と、その `validation`                                   |
| `terraform.tfvars`    | 上の変数の値                                                                        |
| `main.tf`             | モジュールの呼び出し。呼ぶモジュールができてから置く                                |
| `.terraform.lock.hcl` | provider のバージョンとチェックサム。`init` と `just lock` が作成する。コミットする |

## モジュールのファイル構成

| ファイル       | 置くもの                                                                        |
| -------------- | ------------------------------------------------------------------------------- |
| `versions.tf`  | `terraform { required_version, required_providers }`。provider の設定は書かない |
| `variables.tf` | 入力変数。**すべてに `description` と `type`**、任意のものだけ `default`        |
| `main.tf`      | `locals` とリソース定義                                                         |
| `outputs.tf`   | 出力値。**すべてに `description`**                                              |

- `main.tf` が肥大化するときは役割ごとに `<役割>.tf` へ分け (`api.tf` `payment_service.tf` など)、`main.tf` には `locals` と複数の役割で共有する部品だけを残す。1 ファイル 1 つの関心にし、区切り線のコメントでセクションを設けない
- Terraform 内のリソース識別子は役割名にする (`google_cloud_run_v2_service.api`、`google_pubsub_topic.purchase_created`)。同種が 1 つでも `this` にしない
- 使わない変数・出力を定義しない。他モジュールが必要とする ID・名前・URL は出力し、派生形は使う側で組み立てる

## 命名と値の渡し方

- GCP プロジェクトを環境ごとに分けるため、リソース名に環境の接頭辞を付けない (`api`、`payment-service`)。プロジェクト外で一意にする必要があるもの (GCS バケットなど) だけ、先頭にプロジェクト ID を付けて `${var.project_id}-<役割>` にする (`stg-aozora-park-tfstate`)
- リソース名はモジュール内で組み立てる。環境ディレクトリは値 (`project_id` `region` `env`) を渡すだけで名前を組み立てない
- `env` は `validation` で `stg` と `prd` に限る。`project_id` は `validation` で GCP のプロジェクト ID の形式を検査する
- **API キー・シークレット・Webhook URL・メールアドレスをファイルに書かない。** 秘匿値は Secret Manager に置き、Terraform ではシークレットの入れ物と参照だけを定義する。値の投入はユーザーが手で行う
- `terraform.tfvars` には `project_id` `region` `env` のような公開してよい値だけを置く
- モジュール間の受け渡しは outputs 経由で行う。**`module` ブロックに `depends_on` を書かない** (モジュール全体の依存になり循環する)。`for_each` のキーに他モジュールの output を使わない
- ラベルは環境ディレクトリの provider `default_labels` (`project` `environment` `managed_by`) で一括付与し、モジュール内で `labels` を書かない

## 書き方

- `terraform fmt` の整形に従う (2 スペースインデント、連続する引数の `=` を揃える、ブロック間は空行 1 つ)。`just fmt-check` を通す
- 非推奨の書き方をしない。Cloud Run は `google_cloud_run_service` ではなく `google_cloud_run_v2_service` / `google_cloud_run_v2_job` を使う。IAM は `google_*_iam_policy` (全置換) ではなく `google_*_iam_member` を使う。provider のバージョンに対する正しい書き方はドキュメントで確認する
- `google-beta` は beta 限定の引数が要るときだけ使い、その理由をコメントに残す
- バージョンは `~>` で固定する (環境側は `required_version` と provider を固定、モジュール側は下限だけで足りる)
- `required_version` は `.tool-versions` の terraform と同じ系列にする。terraform を上げたら両方を修正する
- `.terraform.lock.hcl` には、手元 (`darwin_arm64`) と CI (`linux_amd64`) の両方のチェックサムを `just lock <env>` で載せる。`init` は実行したマシンの分しか記録しないため。provider のバージョンを上げたときも実行する

## コメント

- 「なぜ」が自明でないときだけ書く。「何をしているか」は書かない
- 日本語で書き、句点 (。) を含めない。文の途中で改行せず 1 行で書く。長くなるなら 1 行 1 論点で行を分ける
- 設計上の判断 (DLQ を付けない理由、min_instances を 0 にする理由、ingress を internal にする理由) は、**理由をリソースの直前にコメントで残す**
- `description` (変数・出力) は英語 1 文で、Terraform のドキュメント慣習に合わせる。コメントは日本語、`description` は英語と使い分ける
- 外部ツールや設計ドキュメントの所在をコメントに書かない

## 実行

- **terraform・tflint・trivy のコマンドはユーザーだけが実行する。** `just` 経由でも同じ。Claude はファイルの作成と調査だけを行い、実行が必要なコマンドを並べて依頼する
- PR を出す前に `just fmt-check` `just lint` と、環境ごとの `just init <env>` `just validate <env>` `just scan <env>` を通す。stg は `just plan` で意図した差分だけが出ることを確かめる
- `apply` と `destroy` のレシピは環境名を省略できない形にする。`prd` には使わない
- trivy は環境ごとに、その環境の `terraform.tfvars` だけを `--tf-vars` で渡す。まとめて渡すと、環境の間で同名の変数が上書きし合う
- backend を変更した直後の `init` は `-reconfigure` を使う (旧 backend に state が無いことを確認したうえで)
- Cloud Run のイメージタグは deploy ワークフローが差し替えるため、Terraform は初回作成と設定 (環境変数・シークレット参照・SA・スケール) だけを担う。`image` は `lifecycle { ignore_changes = [...] }` で無視し、CI の差し替えを drift にしない
- justfile にシェルの処理を書かない (`.claude/rules/general/justfile.md`)
