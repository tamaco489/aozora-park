terraform {
  backend "gcs" {
    bucket = "stg-aozora-park-tfstate"
  }
}
