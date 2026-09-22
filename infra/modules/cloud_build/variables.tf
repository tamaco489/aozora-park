variable "project_id" {
  description = "The ID of the Google Cloud project."
  type        = string
}

variable "region" {
  description = "The region for the connection, the repository link and the trigger."
  type        = string
}

variable "artifact_registry_repository_id" {
  description = "The ID of the Artifact Registry repository the deployer pushes images to."
  type        = string
}

variable "api_service_name" {
  description = "The name of the Cloud Run service for the api that the deployer deploys to."
  type        = string
}

variable "api_service_account_email" {
  description = "The email of the service account the api runs as."
  type        = string
}

variable "enable_tag_trigger" {
  description = "Whether to create the trigger that deploys the api on a tag push."
  type        = bool
}
