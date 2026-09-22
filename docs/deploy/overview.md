# Deployment architecture

[English](./overview.md) | [日本語](./overview.ja.md)

Back to the [documentation index](../README.md).

This page describes how the api reaches Cloud Run. The procedure itself is in [Deploying to stg](./stg.md).
The rules live in the "CD" section of `.claude/rules/ci/coding.md`; this page records the current shape and why it was chosen.

## The deployment paths

![Deployment paths of the api](./images/deploy-flow.png)

Cloud Build runs the build and the deployment inside GCP. The source is fetched from GitHub through Developer Connect, so your local working tree plays no part in it.

## stg and prd

| Item          | stg                                       | prd                                                  |
| ------------- | ----------------------------------------- | ---------------------------------------------------- |
| Started by    | `just deploy-stg <ref>` from your machine | Pushing a tag shaped like `api/v1.2.3`               |
| Trigger       | None                                      | A Cloud Build trigger, defined in Terraform          |
| Approval      | Not required                              | Required                                             |
| Default ref   | `main`                                    | The commit the tag points at                         |
| Image tag     | The commit SHA                            | The commit SHA                                       |
| Current state | Running                                   | The GCP project does not exist yet; definitions only |

stg has no trigger because Developer Connect repositories do not support manual triggers.
Builds are submitted directly with `gcloud builds submit` instead.

Tags are shaped like `api/v1.2.3` so that triggers can be split per service as more services appear.
A tag name contains a slash and cannot be used as an image tag, so images are tagged with the commit SHA.

## The resources involved

| Resource                        | Role                                                                       |
| ------------------------------- | -------------------------------------------------------------------------- |
| Developer Connect connection    | The connection to GitHub. The OAuth token is stored in Secret Manager      |
| Git repository link             | Points at one repository under the connection; Cloud Build fetches from it |
| Cloud Build                     | Runs build → push → deploy as defined in `backend/cloudbuild.yaml`         |
| `sa-deployer`                   | The service account the build and the deployment run as                    |
| Artifact Registry `aozora-park` | Stores the images and keeps the five most recent versions                  |
| Cloud Run `api`                 | Runs the api as `sa-api`                                                   |

The roles of `sa-deployer` are kept to what the deployment needs.

| Role                                       | Granted on                   |
| ------------------------------------------ | ---------------------------- |
| `roles/artifactregistry.writer`            | The `aozora-park` repository |
| `roles/run.developer`                      | The Cloud Run service `api`  |
| `roles/iam.serviceAccountUser`             | `sa-api`                     |
| `roles/developerconnect.readTokenAccessor` | The project                  |
| `roles/logging.logWriter`                  | The project                  |

The last two are granted on the project because Developer Connect connections and links have no resource-level IAM, and log writing cannot be granted below the project.

## What Terraform owns

| Subject                                            | Owned by                                                     |
| -------------------------------------------------- | ------------------------------------------------------------ |
| Cloud Run settings (service account, env, scaling) | Terraform                                                    |
| The Cloud Run `image`                              | The deployment; Terraform ignores it with `ignore_changes`   |
| The Developer Connect connection                   | Created and authorized by hand, then imported into Terraform |
| The OAuth token secret                             | Created by Developer Connect; Terraform does not touch it    |

The connection is created by hand because authorizing GitHub is only possible in a browser. The steps are in the design document, under "5. Developer Connect の接続".

## Alternatives that were not taken

| Alternative                        | Why not                                                                                               |
| ---------------------------------- | ----------------------------------------------------------------------------------------------------- |
| GitHub Actions with WIF            | It hands deployment permissions to GitHub. Cloud Build keeps everything inside GCP                    |
| Deploying automatically on a merge | We want to decide what runs on stg; a deployment that starts with the merge leaves no room to stop it |
| Cloud Deploy                       | With only stg and prd, staged delivery with approvals is more than this project needs                 |
| Artifact Analysis                  | Vulnerability scanning does not pay for itself yet; it can be added when it does                      |
