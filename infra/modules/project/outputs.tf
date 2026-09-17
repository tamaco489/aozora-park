output "project_id" {
  description = "The project ID, available once the required APIs are enabled."
  value       = one(distinct([for s in google_project_service.enabled : s.project]))
}
