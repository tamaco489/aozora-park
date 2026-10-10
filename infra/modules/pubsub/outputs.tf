output "prioritypass_requested_topic" {
  description = "The name of the topic the api publishes a priority pass request to."
  value       = google_pubsub_topic.prioritypass_requested.name
}

output "push_service_account_email" {
  description = "The email of the service account Pub/Sub uses to call the push endpoints."
  value       = google_service_account.push.email
}
