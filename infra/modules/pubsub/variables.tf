variable "project_id" {
  description = "The ID of the Google Cloud project."
  type        = string
}

variable "region" {
  description = "The region of the Cloud Run services the subscriptions push to."
  type        = string
}

variable "publisher_service_account_email" {
  description = "The email of the service account that publishes to the topics."
  type        = string
}

variable "priority_pass_issuer_service_name" {
  description = "The name of the Cloud Run service that handles the priority pass allocation."
  type        = string
}

variable "priority_pass_issuer_uri" {
  description = "The URL of the Cloud Run service that handles the priority pass allocation."
  type        = string
}
