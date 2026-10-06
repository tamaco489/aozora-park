# Pub/Sub のサービスエージェント
#
# DLQ への退避と、push 用 SA の OIDC トークンの作成を Google 側の主体が行うため、この 2 つを明示的に許可する
# メールアドレスの組み立てにプロジェクト番号が要るので、プロジェクトを引いて取り出す
#
# このエージェントは API を有効にしただけでは作られず、初回利用か明示的な作成で実体ができる
# 存在しないまま IAM を付けると 400 で落ちるため、環境ごとに 1 度だけ手で作成しておく
#
#   gcloud beta services identity create --service=pubsub.googleapis.com --project=<プロジェクト ID>
#
# google_project_service_identity で Terraform に持たせることもできるが、google-beta の provider が要る
# provider を増やさず手作業に寄せる判断をしているため、ここでは参照だけにする
data "google_project" "this" {
  project_id = var.project_id
}

locals {
  pubsub_agent = "serviceAccount:service-${data.google_project.this.number}@gcp-sa-pubsub.iam.gserviceaccount.com"

  # push の受け口のパス、worker 側のルーティングと同じ値にする
  push_path = "/pubsub/push"
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/pubsub_topic
# 優先パスの申込を伝えるトピック、イベント名は過去形にする
resource "google_pubsub_topic" "prioritypass_requested" {
  project = var.project_id
  name    = "prioritypass.requested"
}

# 配信を max_delivery_attempts 回失敗したメッセージの退避先
resource "google_pubsub_topic" "prioritypass_requested_dlq" {
  project = var.project_id
  name    = "prioritypass.requested.dlq"
}

# 退避したメッセージを保持するためだけのサブスクリプション
#
# トピックは購読者がいないとメッセージを捨てるため、DLQ に落ちたものを後から読めなくなる
# 再投入は gcloud で手で行うため、push 先を持たない pull のままにする
resource "google_pubsub_subscription" "prioritypass_requested_dlq_hold" {
  project = var.project_id
  name    = "prioritypass.requested.dlq.hold"
  topic   = google_pubsub_topic.prioritypass_requested_dlq.id

  # 既定の 7 日。気づいてから手で再投入するまでの猶予として十分で、延ばすと保管の費用が増える
  message_retention_duration = "604800s"

  # 購読者がいなくても消さない、DLQ の受け皿そのものが消えると退避の意味が無くなる
  expiration_policy {
    ttl = ""
  }
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/service_account
# Pub/Sub が push 先の Cloud Run を叩くときに名乗る SA
# worker ごとには分けない、名乗る主体が同じで、呼び出し先の制限はサービス単位の run.invoker で行うため
resource "google_service_account" "push" {
  project      = var.project_id
  account_id   = "sa-pubsub-push"
  display_name = "Pub/Sub push"
}

# push 先のサービスだけに呼び出しを許す
resource "google_cloud_run_v2_service_iam_member" "priority_pass_issuer_invoker" {
  project  = var.project_id
  location = var.region
  name     = var.priority_pass_issuer_service_name
  role     = "roles/run.invoker"
  member   = google_service_account.push.member
}

# Pub/Sub が sa-pubsub-push の OIDC トークンを作れるようにする
# これが無いと push の設定そのものが作成時に拒否される
resource "google_service_account_iam_member" "push_token_creator" {
  service_account_id = google_service_account.push.name
  role               = "roles/iam.serviceAccountTokenCreator"
  member             = local.pubsub_agent
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/pubsub_subscription
# 申込を priority-pass-issuer に配送するサブスクリプション
resource "google_pubsub_subscription" "prioritypass_requested_allocate" {
  project = var.project_id
  name    = "prioritypass.requested.allocate"
  topic   = google_pubsub_topic.prioritypass_requested.id

  # 割当は Firestore のトランザクション 1 回で終わるため長く待つ必要がない
  # 短すぎると処理中に再配信が始まるので、コールドスタートを含めて余裕を見た値にする
  ack_deadline_seconds = 60

  push_config {
    push_endpoint = "${var.priority_pass_issuer_uri}${local.push_path}"

    # Cloud Run は既定でサービスの URL を audience として検証する
    oidc_token {
      service_account_email = google_service_account.push.email
      audience              = var.priority_pass_issuer_uri
    }
  }

  # 5 回失敗したら DLQ へ送る、再配信で直らない失敗を無限に繰り返さないため
  dead_letter_policy {
    dead_letter_topic     = google_pubsub_topic.prioritypass_requested_dlq.id
    max_delivery_attempts = 5
  }

  retry_policy {
    minimum_backoff = "10s"
    maximum_backoff = "600s"
  }

  # 呼び出し権限が無い状態で作成すると、最初の配信がすべて失敗して DLQ に落ちる
  depends_on = [google_cloud_run_v2_service_iam_member.priority_pass_issuer_invoker]
}

# DLQ への退避は Pub/Sub のサービスエージェントが行うため、退避先への publish と退避元の subscribe を許す
resource "google_pubsub_topic_iam_member" "dlq_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.prioritypass_requested_dlq.name
  role    = "roles/pubsub.publisher"
  member  = local.pubsub_agent
}

resource "google_pubsub_subscription_iam_member" "allocate_subscriber" {
  project      = var.project_id
  subscription = google_pubsub_subscription.prioritypass_requested_allocate.name
  role         = "roles/pubsub.subscriber"
  member       = local.pubsub_agent
}

# api がこのトピックにだけ publish できるようにする、プロジェクト全体のロールは与えない
resource "google_pubsub_topic_iam_member" "prioritypass_requested_publisher" {
  project = var.project_id
  topic   = google_pubsub_topic.prioritypass_requested.name
  role    = "roles/pubsub.publisher"
  member  = "serviceAccount:${var.publisher_service_account_email}"
}
