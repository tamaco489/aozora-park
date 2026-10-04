variable "project_id" {
  description = "The ID of the Google Cloud project."
  type        = string
}

variable "region" {
  description = "The region of the Cloud Run service that Hosting rewrites forward to."
  type        = string
}

variable "api_service_name" {
  description = "The name of the Cloud Run service that Hosting rewrites forward to."
  type        = string
}

variable "deployer_service_account_email" {
  description = "The email of the Cloud Build service account that cd-backend impersonates when it submits a build."
  type        = string
}
