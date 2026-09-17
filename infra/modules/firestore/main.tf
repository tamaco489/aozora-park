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
