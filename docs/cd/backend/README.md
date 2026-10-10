# Backend deployment architecture

[English](./README.md) | [日本語](./README.ja.md)

Back to the [documentation index](../../README.md).

This page describes how the backend services reach Cloud Run. The procedure itself is in [Deploying the backend to stg](./stg.md).
For an overview that also covers the frontend, see [Deployment architecture](../README.md).
The rules live in `.claude/rules/cd/coding.md`; this page records the current shape and why it was chosen.

## The deployment paths

![Deployment paths of the api](./images/flow.png)

Cloud Build runs the build and the deployment inside GCP. The source is fetched from GitHub through Developer Connect, so your local working tree plays no part in it.

**All GitHub Actions does is start the build.** `cd-backend-stg` impersonates `sa-cd-backend` and submits `gcloud beta builds submit`;
the build and the deployment are still carried out by `sa-deployer` inside GCP. Running `just deploy-stg` locally takes exactly the same path, only started by a different principal.

## stg and prd

| Item               | stg                                                                 | prd                                                  |
| ------------------ | ------------------------------------------------------------------- | ---------------------------------------------------- |
| Automatic trigger  | A push to `main` under `backend/**`                                 | Pushing a tag shaped like `<service>/v1.2.3`         |
| Manual trigger     | `workflow_dispatch`, or `just deploy-stg <ref> [service...]`        | None                                                 |
| Who starts it      | `sa-cd-backend` from `cd-backend-stg`, or you                       | A Cloud Build trigger                                |
| What gets deployed | Always every service, one build each                                | Only the service the tag prefix names                |
| Who runs the build | `sa-deployer`                                                       | `sa-deployer`                                        |
| Approval           | Not required                                                        | Required                                             |
| Ref                | `github.sha`, or the first argument to `just deploy-stg` (required) | The commit the tag points at                         |
| Image tag          | The commit SHA                                                      | The commit SHA                                       |
| Current state      | Running                                                             | The GCP project does not exist yet; definitions only |

stg has no Cloud Build trigger because Developer Connect repositories do not support manual triggers.
Builds are submitted directly with `gcloud builds submit` instead, and GitHub Actions uses the same command.

**The arguments to `gcloud beta builds submit`, and the list of services to deploy, therefore exist in two places:** `.github/workflows/cd-backend-stg.yaml` and `backend/scripts/deploy-stg.sh`.
Changing one means changing the other, and both files carry a comment saying so.

Tags are shaped like `<service>/v1.2.3` so that triggers are split per service.
A tag name contains a slash and cannot be used as an image tag, so images are tagged with the commit SHA.

## stg builds every service by default

Each service gets its own build with a different `_SERVICE`. Two services are not built in one build because
`cloudbuild.yaml` writes the digest to the fixed path `/workspace/image_digest.txt`, which would collide within a single build.

The set of targets is not derived from what changed. Touching `internal/platform` or `go.mod` ends up covering both services anyway,
so such a check would mostly add a way for one service to be silently left behind on an older image.

**Only a deployment started by a person can narrow the targets.** A push to `main` always covers every service.

| How it is started   | How the targets are chosen                                       |
| ------------------- | ---------------------------------------------------------------- |
| Push to `main`      | Not selectable; always every service                             |
| `workflow_dispatch` | A JSON array in `services`; every service when it is left empty  |
| `just deploy-stg`   | Service names after the ref; every service when they are omitted |

**If one build fails, the result of the other is still visible.**

| How it is started | How that works                                                                        |
| ----------------- | ------------------------------------------------------------------------------------- |
| `cd-backend-stg`  | The services are a `strategy.matrix`, with `fail-fast: false` so neither is cancelled |
| `just deploy-stg` | The script submits them in turn, reports every failed service and exits with 1        |

The `concurrency` of the workflow serializes runs against each other; jobs within the same run are not held by it.

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

| The original concern                      | Why it no longer applies                                                                                                                                                                                                                         |
| ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| It hands deployment permissions to GitHub | **It does not.** `sa-cd-backend` only holds `cloudbuild.builds.editor` and log viewing; `sa-deployer` still performs the build and the deployment inside GCP. The credential is a short-lived WIF token, and nothing is stored in GitHub Secrets |
| We want to decide what runs on stg        | **We still can.** `just deploy-stg <ref> [service...]` and `workflow_dispatch` both remain, so something other than `main`, for a chosen set of services, can be placed on stg. What changed is the default, which is now "the same as `main`"   |
