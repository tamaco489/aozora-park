output "workload_identity_provider_name" {
  description = "The full resource name of the provider, passed to google-github-actions/auth."
  value       = google_iam_workload_identity_pool_provider.github.name
}

output "cd_backend_service_account_email" {
  description = "The email of the service account that cd-backend impersonates."
  value       = google_service_account.cd_backend.email
}

output "cd_frontend_service_account_email" {
  description = "The email of the service account that cd-frontend impersonates."
  value       = google_service_account.cd_frontend.email
}
