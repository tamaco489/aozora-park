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
