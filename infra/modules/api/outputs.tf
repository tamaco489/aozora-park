output "service_name" {
  description = "The name of the Cloud Run service for the api."
  value       = google_cloud_run_v2_service.api.name
}

output "uri" {
  description = "The URL of the Cloud Run service for the api."
  value       = google_cloud_run_v2_service.api.uri
}

output "service_account_email" {
  description = "The email of the service account the api runs as."
  value       = google_service_account.api.email
}
