locals {
  # 使うようになった API のみ追加していく
  services = toset([
    "artifactregistry.googleapis.com",
    "firestore.googleapis.com",
    "identitytoolkit.googleapis.com",
  ])
}

resource "google_project_service" "enabled" {
  for_each = local.services

  project = var.project_id
  service = each.value
}
