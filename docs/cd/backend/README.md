# Backend deployment architecture

[English](./README.md) | [日本語](./README.ja.md)

Back to the [documentation index](../../README.md).

This page describes how the backend services reach Cloud Run. The procedure itself is in [Deploying the backend to stg](./stg.md).
For an overview that also covers the frontend, see [Deployment architecture](../README.md).
The rules live in `.claude/rules/cd/coding.md`; this page records the current shape and why it was chosen.

## The deployment paths

![Deployment paths of the api](./images/flow.png)

Cloud Build runs the build and the deployment inside GCP. The source is fetched from GitHub through Developer Connect, so your local working tree plays no part in it.

**All GitHub Actions does is start the build.** A per-service workflow (`cd-api-stg` and so on) impersonates `sa-cd-backend` and submits `gcloud beta builds submit`;
the build and the deployment are still carried out by `sa-deployer` inside GCP. Running `just deploy-api-stg` locally takes exactly the same path, only started by a different principal.

## stg and prd

| Item               | stg                                                                     | prd                                                  |
| ------------------ | ----------------------------------------------------------------------- | ---------------------------------------------------- |
| Automatic trigger  | A push to `main` under `backend/**`                                     | Pushing a tag shaped like `<service>/v1.2.3`         |
| Manual trigger     | A per-service `workflow_dispatch`, or `just deploy-<service>-stg <ref>` | None                                                 |
| Who starts it      | `sa-cd-backend` from `cd-<service>-stg`, or you                         | A Cloud Build trigger                                |
| What gets deployed | One service per workflow; a push to `main` starts all of them           | Only the service the tag prefix names                |
| Who runs the build | `sa-deployer`                                                           | `sa-deployer`                                        |
| Approval           | Not required                                                            | Required                                             |
| Ref                | `github.sha`, or the first argument to the just recipe (required)       | The commit the tag points at                         |
| Image tag          | The commit SHA                                                          | The commit SHA                                       |
| Current state      | Running                                                                 | The GCP project does not exist yet; definitions only |

stg has no Cloud Build trigger because Developer Connect repositories do not support manual triggers.
Builds are submitted directly with `gcloud builds submit` instead, and GitHub Actions uses the same command.

**The same `gcloud beta builds submit` therefore exists in two places:** one script per service under `.github/workflows/scripts/` and under `backend/scripts/`.
Changing one means changing the other, and both files carry a comment saying so.

CI does not call the scripts under `backend/scripts/` because what they receive differs. Locally a branch name such as `main` is resolved to a SHA,
whereas CI passes `github.sha` and has nothing to resolve.

Tags are shaped like `<service>/v1.2.3` so that triggers are split per service.
A tag name contains a slash and cannot be used as an image tag, so images are tagged with the commit SHA.

## One of everything per service

Workflows, scripts and just recipes alike: **each one deploys exactly one service.**

| Layer          | api                                           | priority-pass-issuer                                           |
| -------------- | --------------------------------------------- | -------------------------------------------------------------- |
| Workflow       | `cd-api-stg.yaml`                             | `cd-priority-pass-issuer-stg.yaml`                             |
| Script in CI   | `.github/workflows/scripts/deploy-api-stg.sh` | `.github/workflows/scripts/deploy-priority-pass-issuer-stg.sh` |
| Script locally | `backend/scripts/deploy-api-stg.sh`           | `backend/scripts/deploy-priority-pass-issuer-stg.sh`           |
| just recipe    | `just deploy-api-stg <ref>`                   | `just deploy-priority-pass-issuer-stg <ref>`                   |

There are three reasons for the split.

- **How a deployment works is a per-service matter.** A Cloud Run service and a Cloud Run job need different commands, so a single file would accumulate branches
- **A failure on one side does not take the other down.** They are separate workflows, and their results are reported separately
- **A person can pick a target.** Only the workflow of the service you want to ship is started through `workflow_dispatch`

Two services are not built in one build because `cloudbuild.yaml` writes the digest to the fixed path `/workspace/image_digest.txt`, which would collide within a single build.

**No workflow narrows its `paths` on a push to `main`.** Deriving the target from what changed leaves one service behind on an older image
as soon as `internal/platform` or `go.mod` is touched. A merge that touches `backend/**` therefore starts every service's workflow.

The `concurrency` group is per workflow, so deploying api never waits for priority-pass-issuer.

## The resources involved

| Resource                         | Role                                                                       |
| -------------------------------- | -------------------------------------------------------------------------- |
| Developer Connect connection     | The connection to GitHub. The OAuth token is stored in Secret Manager      |
| Git repository link              | Points at one repository under the connection; Cloud Build fetches from it |
| Cloud Build                      | Runs build → push → deploy as defined in `backend/cloudbuild.yaml`         |
| `sa-deployer`                    | The service account the build and the deployment run as                    |
| Artifact Registry `aozora-park`  | Stores the images and keeps the five most recent versions                  |
| Cloud Run `api`                  | Runs the api as `sa-api`                                                   |
| Cloud Run `priority-pass-issuer` | Runs the priority pass allocation as `sa-priority-pass-issuer`             |

The roles of `sa-deployer` are kept to what the deployment needs.

| Role                                       | Granted on                          |
| ------------------------------------------ | ----------------------------------- |
| `roles/artifactregistry.writer`            | The `aozora-park` repository        |
| `roles/run.developer`                      | The Cloud Run services deployed to  |
| `roles/iam.serviceAccountUser`             | The runtime service account of each |
| `roles/developerconnect.readTokenAccessor` | The project                         |
| `roles/logging.logWriter`                  | The project                         |

The last two are granted on the project because Developer Connect connections and links have no resource-level IAM, and log writing cannot be granted below the project.

## What Terraform owns

| Subject                                            | Owned by                                                     |
| -------------------------------------------------- | ------------------------------------------------------------ |
| Cloud Run settings (service account, env, scaling) | Terraform                                                    |
| The Cloud Run `image`                              | The deployment; Terraform ignores it with `ignore_changes`   |
| The Developer Connect connection                   | Created and authorized by hand, then imported into Terraform |
| The OAuth token secret                             | Created by Developer Connect; Terraform does not touch it    |

The connection is created by hand because authorizing GitHub is only possible in a browser. The steps live in the design notes, in the section titled "5. Developer Connect の接続" (Setting up the Developer Connect connection).

## Alternatives that were not taken

| Alternative                            | Why not                                                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Deploying directly from GitHub Actions | It would hand the build and deployment permissions to GitHub. Letting it only start the build is enough |
| Cloud Deploy                           | With only stg and prd, staged delivery with approvals is more than this project needs                   |
| Artifact Analysis                      | Vulnerability scanning does not pay for itself yet; it can be added when it does                        |

### Two previously rejected options were adopted

"GitHub Actions with WIF" and "deploying automatically on a merge" were both rejected at first.
Each concern has since been addressed.

| The original concern                      | Why it no longer applies                                                                                                                                                                                                                          |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| It hands deployment permissions to GitHub | **It does not.** `sa-cd-backend` only holds `cloudbuild.builds.editor` and log viewing; `sa-deployer` still performs the build and the deployment inside GCP. The credential is a short-lived WIF token, and nothing is stored in GitHub Secrets  |
| We want to decide what runs on stg        | **We still can.** `just deploy-<service>-stg <ref>` and a per-service `workflow_dispatch` both remain, so something other than `main`, for a chosen service, can be placed on stg. What changed is the default, which is now "the same as `main`" |
