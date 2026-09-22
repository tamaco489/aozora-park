module "project" {
  source = "../../modules/project"

  project_id = var.project_id
}

module "artifact_registry" {
  source = "../../modules/artifact_registry"

  # API の有効化が完了してから作成されるよう、project モジュールの出力を渡す
  project_id = module.project.project_id
  region     = var.region
}

module "firestore" {
  source = "../../modules/firestore"

  project_id = module.project.project_id
  region     = var.region
}

module "identity_platform" {
  source = "../../modules/identity_platform"

  project_id = module.project.project_id
}

module "api" {
  source = "../../modules/api"

  project_id = module.project.project_id
  region     = var.region
}

module "cloud_build" {
  source = "../../modules/cloud_build"

  project_id                      = module.project.project_id
  region                          = var.region
  artifact_registry_repository_id = module.artifact_registry.repository_id
  api_service_name                = module.api.service_name
  api_service_account_email       = module.api.service_account_email
  enable_tag_trigger              = false
}
