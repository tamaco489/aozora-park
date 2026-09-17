variable "project_id" {
  description = "The ID of the Google Cloud project."
  type        = string

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{4,28}[a-z0-9]$", var.project_id))
    error_message = "project_id must be a valid Google Cloud project ID."
  }
}

variable "region" {
  description = "The default region for resources."
  type        = string
}

variable "env" {
  description = "The environment name."
  type        = string

  validation {
    condition     = contains(["stg", "prd"], var.env)
    error_message = "env must be one of: stg, prd."
  }
}
