provider "google" {
  project = var.project_id
  region  = var.region

  default_labels = {
    project     = "aozora-park"
    environment = var.env
    managed_by  = "terraform"
  }
}
