provider "google" {
  project = var.project_id
  region  = var.region

  # identitytoolkit などは quota project を要求するため、ユーザーの ADC でも課金先を対象のプロジェクトにする
  user_project_override = true
  billing_project       = var.project_id

  default_labels = {
    project     = "aozora-park"
    environment = var.env
    managed_by  = "terraform"
  }
}
