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

variable "run_services" {
  description = "The Cloud Run services the deployer builds and deploys, keyed by the service name."
  type = map(object({
    name                  = string
    service_account_email = string
  }))
}

variable "enable_tag_trigger" {
  description = "Whether to create the triggers that deploy each service on a tag push."
  type        = bool
}
