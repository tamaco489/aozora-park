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

module "priority_pass_issuer" {
  source = "../../modules/priority_pass_issuer"

  project_id = module.project.project_id
  region     = var.region
}

module "pubsub" {
  source = "../../modules/pubsub"

  project_id                        = module.project.project_id
  region                            = var.region
  publisher_service_account_email   = module.api.service_account_email
  priority_pass_issuer_service_name = module.priority_pass_issuer.service_name
  priority_pass_issuer_uri          = module.priority_pass_issuer.uri
}

module "cloud_build" {
  source = "../../modules/cloud_build"

  project_id                      = module.project.project_id
  region                          = var.region
  artifact_registry_repository_id = module.artifact_registry.repository_id
  enable_tag_trigger              = false

  # キーは Cloud Run のサービス名、prd のタグの接頭辞にもそのまま使う
  run_services = {
    "api" = {
      name                  = module.api.service_name
      service_account_email = module.api.service_account_email
    }
    "priority-pass-issuer" = {
      name                  = module.priority_pass_issuer.service_name
      service_account_email = module.priority_pass_issuer.service_account_email
    }
  }
}

module "github_oidc" {
  source = "../../modules/github_oidc"

  project_id                     = module.project.project_id
  region                         = var.region
  api_service_name               = module.api.service_name
  deployer_service_account_email = module.cloud_build.deployer_email
}
