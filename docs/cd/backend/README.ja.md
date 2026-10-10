# backend のデプロイの構成

[English](./README.md) | [日本語](./README.ja.md)

[ドキュメント一覧](../../README.ja.md)に戻る。

backend のサービスを Cloud Run へ届ける経路をまとめます。手順は [backend の stg へのデプロイ](./stg.ja.md)にあります。
frontend を含めた全体の俯瞰は[デプロイの構成](../README.ja.md)にあります。
規約は `.claude/rules/cd/coding.md` が持ちます。ここには現状と、その形を選んだ理由を置きます。

## デプロイの経路

![api のデプロイ経路](./images/flow.png)

ビルドとデプロイは Cloud Build が GCP の中で実行します。ソースは GitHub から Developer Connect 経由で取得するため、手元の作業ツリーは関係しません。

**GitHub Actions が行うのは Cloud Build の起動だけです。** サービスごとのワークフロー (`cd-api-stg` など) が `sa-cd-backend` を名乗って `gcloud beta builds submit` を投げ、
実際のビルドとデプロイは今までどおり `sa-deployer` が GCP の中で行います。手元から `just deploy-api-stg` を実行した場合と、起動する人が違うだけで経路は同じです。

## stg と prd の違い

| 項目                | stg                                                                             | prd                                     |
| ------------------- | ------------------------------------------------------------------------------- | --------------------------------------- |
| 自動の起点          | `main` への push (`backend/**`)                                                 | `<サービス名>/v1.2.3` の形のタグの push |
| 手で起こす          | サービスごとの `workflow_dispatch`、または `just deploy-<サービス名>-stg <ref>` | 無し                                    |
| 起動する主体        | `cd-<サービス名>-stg` の `sa-cd-backend`、または実行者                          | Cloud Build のトリガ                    |
| デプロイ対象        | ワークフロー 1 本につき 1 サービス。`main` への push では全部が起動する         | タグの接頭辞が指すサービスだけ          |
| ビルドを実行する SA | `sa-deployer`                                                                   | `sa-deployer`                           |
| 承認                | 無し                                                                            | 必須                                    |
| ref                 | `github.sha`、または just のレシピの第 1 引数 (省略できない)                    | タグが指すコミット                      |
| イメージのタグ      | コミットの SHA                                                                  | コミットの SHA                          |
| 現在の状態          | 稼働中                                                                          | GCP のプロジェクトは未作成。定義のみ    |

stg に Cloud Build のトリガを作らないのは、Developer Connect のリポジトリが手動のトリガに対応していないためです。
代わりに `gcloud builds submit` でビルドを直接投げます。GitHub Actions から起こす場合も同じコマンドを使います。

**そのため同じ `gcloud beta builds submit` が 2 か所にあります。** `.github/workflows/scripts/` と `backend/scripts/` に、サービスごとに 1 本ずつ置いています。
片方を変更したらもう片方も直す必要があり、両方のファイルに相互参照のコメントを置いています。

CI 側から `backend/scripts/` を呼ばないのは、渡すものが違うためです。手元は `main` のようなブランチ名を受け取って SHA に解決しますが、
CI は `github.sha` をそのまま使うため解決が要りません。

タグを `<サービス名>/v1.2.3` の形にしたのは、サービスごとにトリガを分けるためです。
タグ名はスラッシュを含みイメージのタグに使えないため、イメージにはコミットの SHA を付けます。

## サービスごとに 1 本ずつ分けています

ワークフローもスクリプトも just のレシピも、**1 本が 1 サービスだけを配信します。**

| 層                | api                                           | priority-pass-issuer                                           |
| ----------------- | --------------------------------------------- | -------------------------------------------------------------- |
| ワークフロー      | `cd-api-stg.yaml`                             | `cd-priority-pass-issuer-stg.yaml`                             |
| CI 側のスクリプト | `.github/workflows/scripts/deploy-api-stg.sh` | `.github/workflows/scripts/deploy-priority-pass-issuer-stg.sh` |
| 手元のスクリプト  | `backend/scripts/deploy-api-stg.sh`           | `backend/scripts/deploy-priority-pass-issuer-stg.sh`           |
| just のレシピ     | `just deploy-api-stg <ref>`                   | `just deploy-priority-pass-issuer-stg <ref>`                   |

分ける理由は 3 つあります。

- **デプロイのしかたが変わるのはサービス単位です。** Cloud Run の service と job では使うコマンドが違うため、1 本にまとめると分岐が増えていきます
- **片方の失敗がもう片方を巻き込みません。** 別のワークフローなので、結果も別々に見えます
- **手で起こすときに対象を選べます。** 出したいサービスのワークフローだけを `workflow_dispatch` で起こします

1 回のビルドで 2 サービスを作る形にしないのは、`cloudbuild.yaml` がダイジェストを書く `/workspace/image_digest.txt` が固定名で、同じビルドの中では衝突するためです。

**`main` への push では、どのサービスも `paths` を絞りません。** 変更の内容から対象を判定すると、`internal/platform` や `go.mod` を触ったときに片方が古いまま残ります。
結果として `backend/**` を触るマージでは全サービスのワークフローが起動します。

`concurrency` の group はワークフローごとに分かれているため、api の配信が priority-pass-issuer の配信を待つことはありません。

## 登場するリソース

| リソース                         | 役割                                                                    |
| -------------------------------- | ----------------------------------------------------------------------- |
| Developer Connect の接続         | GitHub との接続。OAuth トークンは Secret Manager に保存される           |
| git repository link              | 接続の下でリポジトリ 1 つを指す。Cloud Build はここからソースを取得する |
| Cloud Build                      | `backend/cloudbuild.yaml` に沿って build → push → deploy を実行する     |
| `sa-deployer`                    | ビルドとデプロイの実行 SA                                               |
| Artifact Registry `aozora-park`  | イメージの置き場所。最新 5 世代を残す                                   |
| Cloud Run `api`                  | api の実行環境。実行 SA は `sa-api`                                     |
| Cloud Run `priority-pass-issuer` | 優先パスの割当の実行環境。実行 SA は `sa-priority-pass-issuer`          |

`sa-deployer` の権限は必要な範囲に絞っています。

| ロール                                     | 付与先                            |
| ------------------------------------------ | --------------------------------- |
| `roles/artifactregistry.writer`            | リポジトリ `aozora-park`          |
| `roles/run.developer`                      | デプロイ対象の Cloud Run サービス |
| `roles/iam.serviceAccountUser`             | 各サービスの実行 SA               |
| `roles/developerconnect.readTokenAccessor` | プロジェクト                      |
| `roles/logging.logWriter`                  | プロジェクト                      |

最後の 2 つをプロジェクトに付けているのは、Developer Connect の接続とリンクがリソース単位の IAM を持たず、ログの書き込みもプロジェクトより下の単位で付けられないためです。

## Terraform との分担

| 対象                                      | 管理するもの                                       |
| ----------------------------------------- | -------------------------------------------------- |
| Cloud Run の設定 (SA・環境変数・スケール) | Terraform                                          |
| Cloud Run の `image`                      | デプロイ。Terraform は `ignore_changes` で無視する |
| Developer Connect の接続                  | 手作業で作成して認可し、Terraform に import する   |
| OAuth トークンのシークレット              | Developer Connect が作成する。Terraform は触らない |

接続を手作業で作るのは、GitHub の認可がブラウザでしか行えないためです。作成の手順は設計ドキュメントの「5. Developer Connect の接続」にあります。

## 採らなかった案

| 案                                | 採らなかった理由                                                             |
| --------------------------------- | ---------------------------------------------------------------------------- |
| GitHub Actions からの直接デプロイ | ビルドとデプロイの権限を GitHub 側に出すことになる。起動だけを任せれば足りる |
| Cloud Deploy                      | 環境が stg と prd の 2 つで、承認付きの段階的な配信までは要らない            |
| Artifact Analysis                 | 脆弱性スキャンは費用に見合う段階ではない。必要になった時点で入れる           |

### 一度は見送った 2 案を採りました

当初は「GitHub Actions + WIF」と「`main` へのマージで自動デプロイ」のどちらも採りませんでした。
懸念はそれぞれ次のように解消しています。

| 当初の懸念                                 | 解消した理由                                                                                                                                                                                                                                           |
| ------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| デプロイの権限を GitHub 側に出すことになる | **出していません。** `sa-cd-backend` に与えたのは Cloud Build の起動 (`cloudbuild.builds.editor`) とログの参照だけで、ビルドとデプロイは `sa-deployer` が GCP の中で行います。資格情報も WIF の短命なトークンで、GitHub Secrets には何も置いていません |
| stg に何が載っているかを意図して決めたい   | **決められます。** `just deploy-<サービス名>-stg <ref>` とサービスごとの `workflow_dispatch` があるため、`main` とは違う内容を、サービスを選んで載せられます。そのうえで既定を「`main` と同じ」にしました                                              |
