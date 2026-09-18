output "api_uri" {
  description = "The URL of the Cloud Run service for the api."
  value       = module.api.uri
}
