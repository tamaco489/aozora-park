resource "google_artifact_registry_repository" "app" {
  project       = var.project_id
  location      = var.region
  repository_id = "aozora-park"
  description   = "Container images for Aozora Park services"
  format        = "DOCKER"

  cleanup_policy_dry_run = false

  # 保持のルールは削除のルールより優先されるため、全件を削除対象にして最新 5 世代だけを残す
  cleanup_policies {
    id     = "delete-all"
    action = "DELETE"

    condition {
      tag_state = "ANY"
    }
  }

  cleanup_policies {
    id     = "keep-latest-5"
    action = "KEEP"

    most_recent_versions {
      keep_count = 5
    }
  }
}
