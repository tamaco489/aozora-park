output "api_uri" {
  description = "The URL of the Cloud Run service for the api."
  value       = module.api.uri
}

output "priority_pass_issuer_uri" {
  description = "The URL of the Cloud Run service for the priority pass issuer."
  value       = module.priority_pass_issuer.uri
}

output "prioritypass_requested_topic" {
  description = "The name of the topic the api publishes a priority pass request to."
  value       = module.pubsub.prioritypass_requested_topic
}

output "git_repository_link_name" {
  description = "The resource name of the git repository link that Cloud Build fetches the source from."
  value       = module.cloud_build.git_repository_link_name
}

output "deployer_email" {
  description = "The email of the service account that runs the builds and deployments."
  value       = module.cloud_build.deployer_email
}

output "workload_identity_provider_name" {
  description = "The full resource name of the provider, passed to google-github-actions/auth."
  value       = module.github_oidc.workload_identity_provider_name
}

output "cd_backend_service_account_email" {
  description = "The email of the service account that cd-backend impersonates."
  value       = module.github_oidc.cd_backend_service_account_email
}

output "cd_frontend_service_account_email" {
  description = "The email of the service account that cd-frontend impersonates."
  value       = module.github_oidc.cd_frontend_service_account_email
}
