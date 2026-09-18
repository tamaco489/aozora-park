# sa-deployer に付けるプロジェクトのロール (Developer Connect のリンクのトークン取得、ログの書き込み)
# Developer Connect の接続とリンクはリソース単位の IAM を持たないため、プロジェクトに付ける
locals {
  sa_deployer_roles = toset([
    "roles/developerconnect.readTokenAccessor",
    "roles/logging.logWriter",
  ])
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/developer_connect_connection
# GitHub との接続
# 認可はブラウザでしか行えないため手作業で作成して import する
# インストール ID と OAuth トークンのシークレットは認可で決まるため書かず、state の値を保つ
resource "google_developer_connect_connection" "github" {
  project       = var.project_id
  location      = var.region
  connection_id = "github"

  github_config {
    github_app = "DEVELOPER_CONNECT"
  }
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/developer_connect_git_repository_link
# Cloud Build がソースを取得するリポジトリ
resource "google_developer_connect_git_repository_link" "aozora_park" {
  project                = var.project_id
  location               = var.region
  parent_connection      = google_developer_connect_connection.github.connection_id
  git_repository_link_id = "aozora-park"
  clone_uri              = "https://github.com/tamaco489/aozora-park.git"
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account
# Cloud Build のビルドとデプロイが名乗る実行 SA
resource "google_service_account" "deployer" {
  project      = var.project_id
  account_id   = "sa-deployer"
  display_name = "deployer"
}

# 実行 SA に locals のロールを 1 つずつ付与する
resource "google_project_iam_member" "deployer" {
  for_each = local.sa_deployer_roles

  project = var.project_id
  role    = each.value
  member  = google_service_account.deployer.member
}

# イメージの push 先のリポジトリだけに書き込み権限を付ける
resource "google_artifact_registry_repository_iam_member" "deployer_writer" {
  project    = var.project_id
  location   = var.region
  repository = var.artifact_registry_repository_id
  role       = "roles/artifactregistry.writer"
  member     = google_service_account.deployer.member
}

# api のサービスだけにリビジョンのデプロイを許可する
# IAM ポリシーは Terraform が管理するため、変更できる run.admin は付けない
resource "google_cloud_run_v2_service_iam_member" "deployer_developer" {
  project  = var.project_id
  location = var.region
  name     = var.api_service_name
  role     = "roles/run.developer"
  member   = google_service_account.deployer.member
}

# デプロイするリビジョンに sa-api を名乗らせるため、sa-api への actAs を付ける
resource "google_service_account_iam_member" "deployer_act_as_api" {
  service_account_id = "projects/${var.project_id}/serviceAccounts/${var.api_service_account_email}"
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.deployer.member
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloudbuild_trigger
# api のタグ (api/v1.2.3) の push で、ビルドからデプロイまでを行うトリガ
# Developer Connect のリポジトリは手動のトリガを作れないため、stg は gcloud builds submit で実行し、トリガーは prd だけに作る
# タグ名はスラッシュを含みイメージのタグに使えないため、イメージのタグにはコミットの SHA を渡す
resource "google_cloudbuild_trigger" "api" {
  count = var.enable_tag_trigger ? 1 : 0

  project         = var.project_id
  location        = var.region
  name            = "api"
  description     = "Build and deploy the api on an api/v* tag push"
  service_account = google_service_account.deployer.id
  filename        = "backend/cloudbuild.yaml"

  # タグの push だけで本番のデプロイが走らないよう、ビルドの開始前に承認を必須にする
  approval_config {
    approval_required = true
  }

  developer_connect_event_config {
    git_repository_link = google_developer_connect_git_repository_link.aozora_park.name

    push {
      tag = "^api/v[0-9]+\\.[0-9]+\\.[0-9]+$"
    }
  }

  substitutions = {
    _SERVICE = "api"
    _TAG     = "$COMMIT_SHA"
  }
}
