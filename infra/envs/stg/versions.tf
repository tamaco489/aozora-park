terraform {
  required_version = "~> 1.16.0"

  required_providers {
    # NOTE: https://github.com/hashicorp/terraform-provider-google/releases
    google = {
      source  = "hashicorp/google"
      version = "~> 8.3"
    }
  }
}
