# Deploying the frontend to stg

[English](./stg.md) | [日本語](./stg.ja.md)

Back to the [documentation index](../../README.md).

Update the stg frontend with a single command from your machine. The overall architecture is described in [Frontend deployment architecture](./README.md).
**Both the build and the deployment run locally.** Unlike the backend, Cloud Build is not involved.

## Prerequisites

- firebase-tools is installed. The version is pinned in `.tool-versions`

```sh
asdf plugin add firebase
asdf install firebase 15.32.1
firebase --version
```

> [!NOTE]
> `.tool-versions` lives at the repository root. **The `firebase` command cannot be resolved outside the repository.**

- You are logged in to the firebase CLI. This is separate from `gcloud` application default credentials

```sh
firebase login
firebase projects:list
```

You are ready once `stg-aozora-park` appears in the list marked `(current)`.

- You have permission on Firebase Hosting in `stg-aozora-park`. Project owners already do

## Deploying

```sh
cd frontend
just deploy-stg
```

The recipe depends on `build`, so `npm run build` runs first. **A stale `dist` is never served.**

The steps are as follows.

| Step | Runs on      | What happens                                                            |
| ---- | ------------ | ----------------------------------------------------------------------- |
| 1    | Your machine | `tsc -b` checks types and `vite build` produces `frontend/dist`         |
| 2    | Your machine | `firebase deploy --only hosting --project stg-aozora-park` runs         |
| 3    | Hosting      | Files are uploaded, the version is finalized, and the release goes live |

**Your working tree is what gets served.** Unlike the backend, the source is not fetched from GitHub, so uncommitted changes are included.

`--project` is passed explicitly so that the recipe cannot target a different project if the default in `.firebaserc` ever changes.

## Verification

Open the site.

```sh
open https://stg-aozora-park.web.app
```

> [!NOTE]
> `index.html` is served with `no-cache`, so **a hard reload is not needed to see the new page.**
> If the old page persists, either the deployment did not finish or the `headers` in `firebase.json` are not taking effect.

Check that requests reach the api.

```sh
curl -s -X POST https://stg-aozora-park.web.app/grpc.health.v1.Health/Check \
  -H "Content-Type: application/json" \
  -H "Connect-Protocol-Version: 1" \
  -d '{"service":"aozorapark.park.v1.ParkService"}'
```

It should return `{"status":"SERVING_STATUS_SERVING"}`, which confirms that the Hosting `rewrites` reach Cloud Run.

Confirm that no preflight is sent by watching the Network tab in your browser devtools.
**No `OPTIONS` request at all** is the expected state.

Check the files being served and their `Cache-Control`.

```sh
curl -sD - -o /dev/null https://stg-aozora-park.web.app/ | grep -i cache-control
```

It should return `no-cache`.

## Rollback

Firebase Hosting keeps the release history, so you can return to a previous version.
**No local rebuild is required.**

Open the Hosting page in the Firebase console, find the release list, and choose "Rollback" on the version you want.

```sh
open https://console.firebase.google.com/project/stg-aozora-park/hosting/sites
```

Because `index.html` is served with `no-cache`, the rollback takes effect on the next request.

## How it differs from the backend

| Item                        | Frontend                      | Backend                              |
| --------------------------- | ----------------------------- | ------------------------------------ |
| Where the build runs        | Your machine (a runner in CI) | Cloud Build (inside GCP)             |
| Where the source comes from | Your working tree             | GitHub (through Developer Connect)   |
| Uncommitted changes         | **Included**                  | Not included                         |
| Deployment target           | Firebase Hosting              | Cloud Run                            |
| Rollback                    | Pick a release in the console | `gcloud run services update-traffic` |
| prd                         | Undecided                     | A tag push of the form `api/v1.2.3`  |

- The steps on this page are for starting a deployment by hand. Anything merged into `main` is deployed by `cd-frontend-stg`, so running them is normally unnecessary
- A deployment from CI builds on the runner, so **uncommitted changes are not included.** The working tree is served as-is only when the deployment is started locally
- `firebase.json` is shared with the emulator settings. Adding `hosting` also started the Hosting emulator, so `docker/firebase-emulator/Dockerfile` now pins it to `--only firestore`
