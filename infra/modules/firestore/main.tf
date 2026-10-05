# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/firestore_database
# 無料枠が適用されるのは (default) データベースだけのため、名前付きのデータベースにしない
# 業務データを失わないよう、GCP 側の削除保護に加えて Terraform の destroy と再作成も止める
resource "google_firestore_database" "default" {
  project                 = var.project_id
  name                    = "(default)"
  location_id             = var.region
  type                    = "FIRESTORE_NATIVE"
  database_edition        = "STANDARD"
  delete_protection_state = "DELETE_PROTECTION_ENABLED"
  deletion_policy         = "PREVENT"
}

# NOTE: https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/firestore_index
# 単一フィールドのインデックスは自動で作られるが、date の等価条件と startTime の並び替えを組み合わせた ListTimeSlots のクエリは複合インデックスがないと失敗する
# ListTimeSlots が引くのは 1 つのアトラクション配下の timeSlots だけで、親をまたいで引く経路がないためクエリスコープは COLLECTION にする
# 全アトラクション横断で特定日の枠を集計する経路ができたら、COLLECTION_GROUP のインデックスをその時点で別に追加する
resource "google_firestore_index" "time_slots_by_date" {
  project     = var.project_id
  database    = google_firestore_database.default.name
  collection  = "timeSlots"
  query_scope = "COLLECTION"

  fields {
    field_path = "date"
    order      = "ASCENDING"
  }

  fields {
    field_path = "startTime"
    order      = "ASCENDING"
  }
}
