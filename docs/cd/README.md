# Deployment architecture

[English](./README.md) | [日本語](./README.ja.md)

Back to the [documentation index](../README.md).

This project has two deployment paths that work in different ways. This page gives an overview of both.
The details are in [Backend deployment architecture](./backend/README.md) and [Frontend deployment architecture](./frontend/README.md).
The mechanism used to authenticate to GCP is described in [Workload Identity Federation](./wif/README.md).

## The two paths

|                             | Backend (one per Cloud Run service)                                     | Frontend                                                               |
| --------------------------- | ----------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| Target                      | Cloud Run                                                               | Firebase Hosting                                                       |
| Workflow                    | `cd-<service>-stg.yaml`, one per service                                | `cd-frontend-stg.yaml`                                                 |
| Where the build runs        | **Cloud Build (inside GCP)**                                            | **The GitHub Actions runner**, or your machine when started locally    |
| Where the source comes from | GitHub (through Developer Connect)                                      | The workflow checkout, or your working tree when started locally       |
| Uncommitted changes         | Not included                                                            | Not included from CI. **Included** when started locally                |
| Automatic trigger (stg)     | A push to `main` under `backend/**`                                     | A push to `main` under `frontend/**`, `firebase.json` or `.firebaserc` |
| Manual trigger (stg)        | `workflow_dispatch`, or `cd backend && just deploy-<service>-stg <ref>` | `workflow_dispatch`, or `cd frontend && just deploy-stg`               |
| Rollback                    | `gcloud run services update-traffic`                                    | Pick a release in the Firebase console                                 |
| prd                         | A tag push of the form `<service>/v1.2.3`, with approval required       | Undecided. A tag shaped like `spa/v1.2.3` is the expected direction    |
| Current state               | stg is running. The prd GCP project does not exist yet                  | stg is running. The prd GCP project does not exist yet                 |

In short: **GCP fetches and builds the backend itself, while the frontend is built elsewhere and uploaded.**
All GitHub Actions does for the backend is start a Cloud Build; both the build and the deployment stay inside GCP.

The backend builds inside GCP because producing an image takes time and the build environment needs to be fixed.
The frontend only produces static files in a few seconds, so it stays on the runner.

## Shared decisions

- **stg is deployed automatically on a push to `main`.** The backend runs one `cd-<service>-stg` per service and the frontend runs `cd-frontend-stg`
- **No long-lived credentials live in the repository or in GitHub Secrets.** GCP access uses Workload Identity Federation: the workflow exchanges the OIDC token it gets from `id-token: write` for a short-lived access token
- **The same deployment can still be started by hand.** The backend has one recipe per service and the frontend has `just deploy-stg`, so a broken CI does not block a release
- **`workflow_dispatch` deploys from any branch.** It exists to try a work-in-progress branch on stg, the same purpose the local recipes serve
- **Traffic from the frontend to the api is forwarded to Cloud Run by Firebase Hosting `rewrites`.** The api is on the same origin, so CORS is not configured. Cloud Run stays open to `allUsers`, because traffic forwarded from Hosting does not count as internal. The details are in [Frontend deployment architecture](./frontend/README.md)
- prd only exists as configuration, because the GCP project has not been created yet

### There is no WIF in `prd`

The Workload Identity Federation pool and provider exist only in `stg`.
prd has no agreed deployment path yet, so **there is nothing concrete to grant access to.**

The stg provider looks only at `assertion.repository` and accepts a token from this repository regardless of branch.
That looseness is deliberate, so that a work-in-progress branch can update stg, and **the same condition must not be carried over to prd.**
When prd is automated, the choice between a tag and a branch, and whether an approval step is required, will be made first, and `attribute_condition` written to match.

## Details

| Document                                                 | Contents                                                      |
| -------------------------------------------------------- | ------------------------------------------------------------- |
| [Backend deployment architecture](./backend/README.md)   | Cloud Build and Developer Connect, and what Terraform owns    |
| [Deploying the backend to stg](./backend/stg.md)         | Procedure, verification, rollback                             |
| [Frontend deployment architecture](./frontend/README.md) | Rewrites and caching, and the alternatives that were rejected |
| [Deploying the frontend to stg](./frontend/stg.md)       | Procedure, verification, rollback                             |

The conventions live in `.claude/rules/cd/coding.md`. These pages record the current shape and why it was chosen.
