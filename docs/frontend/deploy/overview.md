# Frontend deployment architecture

[English](./overview.md) | [日本語](./overview.ja.md)

Back to the [documentation index](../../README.md).

This page describes how the frontend reaches Firebase Hosting. The procedure itself is in [Deploying the frontend to stg](./stg.md).
For an overview that also covers the backend, see [Deployment architecture](../../deploy/overview.md).
The conventions live in the "CD" section of `.claude/rules/ci/coding.md`. This page records the current shape and why it was chosen.

## Deployment path

![Deployment path of the frontend](./images/flow.png)

**Both the build and the deployment run on your machine.** Unlike the backend, Cloud Build is not involved.
Only the static files under `frontend/dist` are served; there is no server process.

## Putting the api on the same origin

Firebase Hosting `rewrites` forward RPC paths to the `api` service on Cloud Run.
From the browser's point of view the api is on the same origin, so **no preflight (`OPTIONS`) is sent.**

| Request path       | Destination                                                        |
| ------------------ | ------------------------------------------------------------------ |
| `/aozorapark.*/*`  | The `api` service on Cloud Run                                     |
| `/grpc.health.*/*` | The `api` service on Cloud Run                                     |
| Anything else      | A static file under `frontend/dist`, falling back to `/index.html` |

Connect sends `Content-Type: application/json` and `Connect-Protocol-Version`, so a cross-origin call always carries a preflight.
That adds a round trip per RPC, and with `min_instance_count = 0` on Cloud Run it can also trigger cold starts.

The `rewrites` use `regex` rather than a glob because an RPC path looks like `/aozorapark.park.v1.ParkService/GetPark`, where the first segment contains dots.
The glob `**` matches any number of path segments, and its behaviour is not well defined when it starts in the middle of a segment.

## Local development uses the same shape

`server.proxy` in `frontend/vite.config.ts` forwards the same two prefixes to `http://localhost:8080`.
Local development is therefore same-origin as well, and no preflight is sent.

The base URL is fixed to `/`. It is not switched through an environment variable.
Holding a setting that can only ever take one value would mean putting `.env.*` back under version control, which creates a place where secrets could be added later.

## Caching

Firebase Hosting applies `Cache-Control: max-age=3600` to static content by default.
The `headers` block in `firebase.json` overrides it.

| Target     | Value               | Reason                                                                                                                           |
| ---------- | ------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| Everything | `no-cache`          | Deployments take effect immediately. The response may be stored but must be revalidated, so an unchanged file costs only a `304` |
| RPC        | `private, no-store` | Keeps per-user data out of the CDN                                                                                               |

Revalidating `index.html` is the key part.
A stale `index.html` references hashed filenames that no longer exist in the new release, which breaks the page.

Assets under `/assets/**` are deliberately **not** given a long cache.
`headers` are matched against the request URL **before** rewrites are applied, so a request for a missing asset falls back to `index.html` and that HTML would be pinned for the long duration.
Adding a long cache requires first excluding `/assets/` from the SPA fallback.

## stg versus prd

| Item          | stg                                 | prd                                                                 |
| ------------- | ----------------------------------- | ------------------------------------------------------------------- |
| Trigger       | `just deploy-stg` from your machine | Undecided. A tag shaped like `spa/v1.2.3` is the expected direction |
| Where it runs | Your machine                        | Undecided                                                           |
| Build trigger | None                                | Undecided                                                           |
| Approval      | None                                | Undecided                                                           |
| Current state | Running                             | The GCP project does not exist yet                                  |

Backend tags are shaped like `api/v1.2.3`, picked up by a Cloud Build trigger matching `^api/v[0-9]+\.[0-9]+\.[0-9]+$`.
The prefix exists so that triggers can be split as more deployment targets appear, so the frontend will use `spa/v1.2.3`.

However, **Firebase Hosting has no equivalent of a Cloud Build trigger.**
Serving on a tag push requires CI (GitHub Actions), so how prd is served will be decided together with automating CD.

## Resources involved

| Resource              | Role                                                                                                                                                                                |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Firebase Hosting site | The default site `stg-aozora-park`, created automatically when the project became a Firebase project (which happened as a side effect of enabling Identity Platform in milestone 3) |
| `firebase.json`       | Public directory, `rewrites`, `headers`. Shares the file with the emulator settings                                                                                                 |
| `.firebaserc`         | The default project, so running `firebase` directly cannot target the wrong one by accident                                                                                         |
| firebase-tools        | Managed through `.tool-versions`. Used only for deployment                                                                                                                          |

There are two default domains, and both point at the same site.

| Domain                            | Purpose                                       |
| --------------------------------- | --------------------------------------------- |
| `stg-aozora-park.web.app`         | Serving the site                              |
| `stg-aozora-park.firebaseapp.com` | Used by Identity Platform for OAuth redirects |

## What Terraform owns

Terraform only enables the APIs.

| Target                                      | Owned by                              | Reason                                                                                 |
| ------------------------------------------- | ------------------------------------- | -------------------------------------------------------------------------------------- |
| Enabling `firebase` and `firebasehosting`   | Terraform (`modules/project`)         | `google_project_service` is enough                                                     |
| Turning the project into a Firebase project | Manual                                | `google_firebase_project` requires `google-beta`                                       |
| The default Hosting site                    | Manual                                | Same as above. It is created automatically when the project becomes a Firebase project |
| Hosting configuration and serving           | `firebase.json` and `firebase deploy` | The configuration changes often                                                        |

Adding `google-beta` for only these two resources is not worth it, so they are handled by hand.
The Developer Connect connection is treated the same way.

## Alternatives that were rejected

### Configure CORS and keep separate origins

Configure the allowed origins on the api and call `*.run.app` directly from Hosting.

- Connect caches preflights per URL, so `Access-Control-Max-Age` still leaves one extra round trip per RPC method
- Putting both behind an external load balancer with Cloud Armor costs roughly 18 USD per month in fixed fees

`rewrites` cost nothing extra and remove the preflight entirely.

### Deploy from GitHub Actions (deferred for now)

The design includes `cd-frontend` (running `firebase deploy` on a push to main), but it is not implemented yet.
Reaching GCP from GitHub Actions requires Workload Identity Federation, and setting that up is a prerequisite shared with `cd-infra`.

The token issued by `firebase login:ci` is deprecated; `GOOGLE_APPLICATION_CREDENTIALS` (ADC) is the recommended path.
Pointing that at a WIF credential should avoid long-lived credentials entirely, but this is **not verified yet**.

### Install firebase-tools as a frontend devDependency

Pinning the version through `package-lock.json` is attractive, but it grows the dependency tree from 34 to 700 packages, adds 12 vulnerabilities, and slows `npm ci` in CI.
That is too high a price for a tool used only at deploy time, so it is managed through `.tool-versions` instead.
