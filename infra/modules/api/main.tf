locals {
  sa_api_roles = toset([
    "roles/cloudtrace.agent",
    "roles/datastore.user",
    "roles/logging.logWriter",
  ])
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/google_service_account
resource "google_service_account" "api" {
  project      = var.project_id
  account_id   = "sa-api"
  display_name = "api"
}

resource "google_project_iam_member" "api" {
  for_each = local.sa_api_roles

  project = var.project_id
  role    = each.value
  member  = google_service_account.api.member
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_v2_service
# 認証はアプリ内で行う設計のため、ingress を全トラフィックにして外部から直接受ける
# ステートレスで失うデータが無いため、Terraform の削除保護を外して destroy と再作成をできるようにする
resource "google_cloud_run_v2_service" "api" {
  project             = var.project_id
  name                = "api"
  location            = var.region
  ingress             = "INGRESS_TRAFFIC_ALL"
  deletion_protection = false

  template {
    service_account = google_service_account.api.email

    # コールドスタートを許容して、リクエストが無い間の待機費用をゼロにするため min を 0 にする
    # 想定する負荷は小さく、アクセスの急増や誤ったループで費用が膨らまないよう max を 2 に抑える
    scaling {
      min_instance_count = 0
      max_instance_count = 2
    }

    containers {
      # 初回の作成だけに使う公開サンプル、以降はデプロイが差し替える
      image = "us-docker.pkg.dev/cloudrun/container/hello"

      resources {
        limits = {
          cpu    = "1"
          memory = "512Mi"
        }
        # resources を書くと既定の true が効かなくなるため、リクエストの処理中だけ CPU を割り当てる設定を明示する
        cpu_idle = true
      }

      env {
        name  = "GOOGLE_CLOUD_PROJECT"
        value = var.project_id
      }
    }
  }

  lifecycle {
    # image はデプロイが差し替え、client と client_version はデプロイした gcloud の値に書き換わるため、drift にしない
    ignore_changes = [
      client,
      client_version,
      template[0].containers[0].image,
    ]
  }
}

# 認証はアプリ内で行うため、Cloud Run の IAM による呼び出し元の制限はかけずに誰でも呼べるようにする
resource "google_cloud_run_v2_service_iam_member" "api_invoker" {
  project  = google_cloud_run_v2_service.api.project
  location = google_cloud_run_v2_service.api.location
  name     = google_cloud_run_v2_service.api.name
  role     = "roles/run.invoker"
  member   = "allUsers"
}
