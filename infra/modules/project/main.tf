locals {
  services = toset([
    "artifactregistry.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "firestore.googleapis.com",
    "identitytoolkit.googleapis.com",
  ])
}

resource "google_project_service" "enabled" {
  for_each = local.services

  project = var.project_id
  service = each.value
}
