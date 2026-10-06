# priority-pass-issuer の実行 SA に付けるプロジェクトのロール (トレースの送信、Firestore の読み書き、ログの書き込み)
# 自分からは publish しないため pubsub.publisher は持たない
locals {
  sa_priority_pass_issuer_roles = toset([
    "roles/cloudtrace.agent",
    "roles/datastore.user",
    "roles/logging.logWriter",
  ])
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account
# priority-pass-issuer の Cloud Run が名乗る実行 SA
resource "google_service_account" "priority_pass_issuer" {
  project      = var.project_id
  account_id   = "sa-priority-pass-issuer"
  display_name = "priority-pass-issuer"
}

# 実行 SA に locals のロールを 1 つずつ付与する
resource "google_project_iam_member" "priority_pass_issuer" {
  for_each = local.sa_priority_pass_issuer_roles

  project = var.project_id
  role    = each.value
  member  = google_service_account.priority_pass_issuer.member
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/cloud_run_v2_service
# 優先パスの割当を行う Cloud Run のサービス
# 呼び出すのは Pub/Sub の push だけで外から直接叩く経路がないため、ingress を内部に閉じる
# 呼び出し権限は pubsub モジュールが sa-pubsub-push にだけ与える
# ステートレスで失うデータが無いため、Terraform の削除保護を外して destroy と再作成をできるようにする
resource "google_cloud_run_v2_service" "priority_pass_issuer" {
  project             = var.project_id
  name                = "priority-pass-issuer"
  location            = var.region
  ingress             = "INGRESS_TRAFFIC_INTERNAL_ONLY"
  deletion_protection = false

  template {
    service_account = google_service_account.priority_pass_issuer.email
    scaling {
      min_instance_count = 0 # コールドスタートを許容して、メッセージが無い間の待機費用をゼロにするため min を 0 にする
      max_instance_count = 2 # 想定する負荷は小さく、再配信の繰り返しで費用が膨らまないよう max を 2 に抑える
    }

    containers {
      # 初回の作成だけに使う公開サンプル、以降はデプロイが差し替える
      image = "us-docker.pkg.dev/cloudrun/container/hello"

      # Pub/Sub の push は HTTP/1.1 の POST で届くため、api のような h2c の指定はしない
      ports {
        container_port = 8080
      }

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
