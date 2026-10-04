variable "project_id" {
  description = "The ID of the Google Cloud project."
  type        = string
}

variable "deployer_service_account_email" {
  description = "The email of the Cloud Build service account that cd-backend impersonates when it submits a build."
  type        = string
}
