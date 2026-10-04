# Deployment architecture

[English](./overview.md) | [日本語](./overview.ja.md)

Back to the [documentation index](../README.md).

This project has two deployment paths that work in different ways. This page gives an overview of both.
The details are in [Backend deployment architecture](../backend/deploy/overview.md) and [Frontend deployment architecture](../frontend/deploy/overview.md).

## The two paths

|                      | Backend (api)                                               | Frontend                               |
| -------------------- | ----------------------------------------------------------- | -------------------------------------- |
| Target               | Cloud Run                                                   | Firebase Hosting                       |
| Where the build runs | **Cloud Build (inside GCP)**                                | **Your machine**                       |
| Source of truth      | GitHub (through Developer Connect)                          | Your working tree                      |
| Uncommitted changes  | Not included                                                | **Included**                           |
| Trigger (stg)        | `cd backend && just deploy-stg <ref>`                       | `cd frontend && just deploy-stg`       |
| Rollback             | `gcloud run services update-traffic`                        | Pick a release in the Firebase console |
| prd                  | A tag push of the form `api/v1.2.3`, with approval required | Undecided                              |

In short: **the backend builds what is on GitHub inside GCP, while the frontend builds what is on your machine and uploads it.**

The backend builds inside GCP because producing an image takes time and the build environment needs to be fixed.
The frontend only produces static files in a few seconds, so it stays local.

## Shared decisions

- **Deployment does not go through GitHub Actions today.** Actions only runs checks. Automated deployment (`cd-frontend` and `cd-infra`) is in the design but not implemented, and requires Workload Identity Federation first
- Even once automated, no long-lived credentials will live in the repository or in GitHub Secrets. GCP access will use WIF
- stg is deployed with one command from a developer machine. Both use a recipe named `just deploy-stg`
- prd only exists as configuration, because the GCP project has not been created yet

## How the frontend reaches the api

Traffic from the frontend to the api is **forwarded to Cloud Run by Firebase Hosting `rewrites`.**

```text
Browser ──> Firebase Hosting ──rewrites──> Cloud Run (api)
                   │
                   └── Static files (frontend/dist)
```

From the browser's point of view the api is on the same origin, so no preflight (`OPTIONS`) is sent.
CORS is not configured; the reasoning is in [Frontend deployment architecture](../frontend/deploy/overview.md).

Cloud Run stays open to `allUsers`, because traffic forwarded from Hosting does not count as internal.

## The diagrams

Each diagram lives with the area it describes. Both are shown here side by side.

### Backend

![Deployment paths of the api](../backend/deploy/images/flow.png)

### Frontend

![Deployment path of the frontend](../frontend/deploy/images/flow.png)

## Details

| Document                                                           | Contents                                                      |
| ------------------------------------------------------------------ | ------------------------------------------------------------- |
| [Backend deployment architecture](../backend/deploy/overview.md)   | Cloud Build and Developer Connect, and what Terraform owns    |
| [Deploying the backend to stg](../backend/deploy/stg.md)           | Procedure, verification, rollback                             |
| [Frontend deployment architecture](../frontend/deploy/overview.md) | Rewrites and caching, and the alternatives that were rejected |
| [Deploying the frontend to stg](../frontend/deploy/stg.md)         | Procedure, verification, rollback                             |

The conventions live in the "CD" section of `.claude/rules/ci/coding.md`. These pages record the current shape and why it was chosen.
