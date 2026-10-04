# GitHub Actions から GCP に入るための OIDC の連携先
# リポジトリは cloud_build の clone_uri と同じものに固定する
locals {
  github_repository = "tamaco489/aozora-park"
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/iam_workload_identity_pool
resource "google_iam_workload_identity_pool" "github" {
  project                   = var.project_id
  workload_identity_pool_id = "github"
  display_name              = "github"
  description               = "Identities for GitHub Actions workflows"
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/iam_workload_identity_pool_provider
# 公開のプロバイダのため attribute_condition が必須、無いと任意のリポジトリのトークンを受け入れてしまう
# ref で絞らないのは、ワークフローを main へマージする前に別のブランチから試せるようにするため
resource "google_iam_workload_identity_pool_provider" "github" {
  project                            = var.project_id
  workload_identity_pool_id          = google_iam_workload_identity_pool.github.workload_identity_pool_id
  workload_identity_pool_provider_id = "github"
  display_name                       = "github"
  attribute_condition                = "assertion.repository == '${local.github_repository}'"

  attribute_mapping = {
    "google.subject"       = "assertion.sub"
    "attribute.repository" = "assertion.repository"
  }

  oidc {
    issuer_uri = "https://token.actions.githubusercontent.com"
  }
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account
# cd-backend が名乗る SA、Cloud Build を起動するだけでビルドとデプロイは sa-deployer が行う
resource "google_service_account" "cd_backend" {
  project      = var.project_id
  account_id   = "sa-cd-backend"
  display_name = "cd-backend"
}

# ビルドの作成と、作成したビルドの状態の参照
resource "google_project_iam_member" "cd_backend_builds_editor" {
  project = var.project_id
  role    = "roles/cloudbuild.builds.editor"
  member  = google_service_account.cd_backend.member
}

# ビルドのログを流すために要る、Cloud Build が CLOUD_LOGGING_ONLY でログを出すため
resource "google_project_iam_member" "cd_backend_logging_viewer" {
  project = var.project_id
  role    = "roles/logging.viewer"
  member  = google_service_account.cd_backend.member
}

# ビルドを sa-deployer で走らせるため、sa-deployer への actAs を付ける
resource "google_service_account_iam_member" "cd_backend_act_as_deployer" {
  service_account_id = "projects/${var.project_id}/serviceAccounts/${var.deployer_service_account_email}"
  role               = "roles/iam.serviceAccountUser"
  member             = google_service_account.cd_backend.member
}

# cd-frontend が名乗る SA、Hosting への配信だけを行う
resource "google_service_account" "cd_frontend" {
  project      = var.project_id
  account_id   = "sa-cd-frontend"
  display_name = "cd-frontend"
}

# Hosting はサイト単位の IAM を持たないためプロジェクトに付ける
resource "google_project_iam_member" "cd_frontend_hosting_admin" {
  project = var.project_id
  role    = "roles/firebasehosting.admin"
  member  = google_service_account.cd_frontend.member
}

# Hosting は rewrites の転送先が実在するかを version の確定時に検証するため、api を参照できる必要がある
# プロジェクト全体ではなく api のサービスだけに絞る
resource "google_cloud_run_v2_service_iam_member" "cd_frontend_api_viewer" {
  project  = var.project_id
  location = var.region
  name     = var.api_service_name
  role     = "roles/run.viewer"
  member   = google_service_account.cd_frontend.member
}

# 2 つの SA を、このリポジトリのワークフローからだけ借りられるようにする
resource "google_service_account_iam_member" "cd_backend_workload_identity_user" {
  service_account_id = google_service_account.cd_backend.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${local.github_repository}"
}

resource "google_service_account_iam_member" "cd_frontend_workload_identity_user" {
  service_account_id = google_service_account.cd_frontend.name
  role               = "roles/iam.workloadIdentityUser"
  member             = "principalSet://iam.googleapis.com/${google_iam_workload_identity_pool.github.name}/attribute.repository/${local.github_repository}"
}
