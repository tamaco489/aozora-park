output "service_name" {
  description = "The name of the Cloud Run service for the priority pass issuer."
  value       = google_cloud_run_v2_service.priority_pass_issuer.name
}

output "uri" {
  description = "The URL of the Cloud Run service for the priority pass issuer."
  value       = google_cloud_run_v2_service.priority_pass_issuer.uri
}

output "service_account_email" {
  description = "The email of the service account the priority pass issuer runs as."
  value       = google_service_account.priority_pass_issuer.email
}
