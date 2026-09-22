output "git_repository_link_name" {
  description = "The resource name of the git repository link that Cloud Build fetches the source from."
  value       = google_developer_connect_git_repository_link.aozora_park.name
}

output "deployer_email" {
  description = "The email of the service account that runs the builds and deployments."
  value       = google_service_account.deployer.email
}
