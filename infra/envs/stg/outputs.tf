output "api_uri" {
  description = "The URL of the Cloud Run service for the api."
  value       = module.api.uri
}

output "git_repository_link_name" {
  description = "The resource name of the git repository link that Cloud Build fetches the source from."
  value       = module.cloud_build.git_repository_link_name
}

output "deployer_email" {
  description = "The email of the service account that runs the builds and deployments."
  value       = module.cloud_build.deployer_email
}
