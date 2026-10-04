# Deploying the backend to stg

[English](./stg.md) | [日本語](./stg.ja.md)

Back to the [documentation index](../../README.md).

After merging into `main`, update the api on stg with a single command from your machine. The overall architecture is described in [Backend deployment architecture](./overview.md).
Cloud Build runs the build and the deployment inside GCP, and fetches the source from GitHub through Developer Connect.

## Prerequisites

- You are logged in to `gcloud`.

```sh
gcloud auth login
```

- You can do both of the following on `stg-aozora-park`. Being a project owner satisfies both.
  - Create builds (`cloudbuild.builds.create`)
  - Use `sa-deployer` (`iam.serviceAccountUser` on `sa-deployer`)
- The commit you want to deploy is pushed to GitHub.

**Local changes are not deployed.** The source comes from GitHub, so anything not committed and pushed is left out of the build.

## Deploy

```sh
cd backend
just deploy-stg              # deploys main
just deploy-stg <ref>        # deploys a branch, a tag or a SHA
```

`<ref>` is resolved as `origin/<ref>` first, and used as given when that does not exist.
The resolved SHA becomes the image tag, so the image tells you which commit is running.

| Step | Runs on      | What happens                                                                            |
| ---- | ------------ | --------------------------------------------------------------------------------------- |
| 1    | Your machine | Resolves the ref to a commit SHA and submits the build with `gcloud beta builds submit` |
| 2    | Cloud Build  | Fetches the source at that SHA through the Developer Connect repository link            |
| 3    | Cloud Build  | Builds the image with `backend/Dockerfile` and pushes it to Artifact Registry           |
| 4    | Cloud Build  | Deploys the pushed image by digest to the Cloud Run service `api`                       |

The build runs as `sa-deployer`. Logs are streamed to the command output.
The history is in the [Cloud Build list](https://console.cloud.google.com/cloud-build/builds?project=stg-aozora-park); select the `asia-northeast1` region.

> [!NOTE]
> The build config (`backend/cloudbuild.yaml`) is read from your working tree.
> Only the source comes from GitHub, so a local change to `cloudbuild.yaml` takes effect immediately.

## Verifying the deployment

Get the URL from the Terraform output.

```sh
cd infra
terraform -chdir=envs/stg output -raw api_uri
```

Call the health check.

```sh
buf curl -d '{"service":""}' <api_uri>/grpc.health.v1.Health/Check
```

It should return `{"status":"SERVING_STATUS_SERVING"}`.

> [!NOTE]
> Without `--schema`, `buf curl` discovers the RPC definitions through server reflection.
> Reflection is gRPC and therefore needs HTTP/2.
> Cloud Run talks HTTP/2 with the client, but converts requests to HTTP/1.1 when forwarding them to the container by default.
> The container port is therefore named `h2c` in Terraform, which keeps HTTP/2 all the way to the container.

Create and read a park. **This writes to the Firestore database of stg.** Delete the documents you create for verification afterwards.

```sh
buf curl -d '{"name":"デプロイ確認","defaultDailyCapacity":100,"inventoryDays":7}' \
  <api_uri>/aozorapark.park.v1.ParkService/CreatePark

buf curl -d '{"parkId":"<the park_id you created>"}' \
  <api_uri>/aozorapark.park.v1.ParkService/GetPark
```

Check which revision and image are serving.

```sh
gcloud run services describe api --region=asia-northeast1 --project=stg-aozora-park \
  --format='value(status.latestReadyRevisionName,spec.template.spec.containers[0].image)'
```

## Rolling back

Send the traffic back to the previous revision. Cloud Run keeps revisions, so there is no image to rebuild.

```sh
gcloud run revisions list --service=api --region=asia-northeast1 --project=stg-aozora-park

gcloud run services update-traffic api \
  --region=asia-northeast1 --project=stg-aozora-park \
  --to-revisions=<the revision to roll back to>=100
```

After rolling back, move the traffic to the newest revision again with the following command, or by deploying again.

```sh
gcloud run services update-traffic api \
  --region=asia-northeast1 --project=stg-aozora-park --to-latest
```

## How it works, and how prd differs

| Item                | stg                                                              | prd                                                      |
| ------------------- | ---------------------------------------------------------------- | -------------------------------------------------------- |
| Automatic trigger   | A push to `main`, which starts `cd-backend-stg`                  | Pushing a tag shaped like `api/v1.2.3`                   |
| Manual trigger      | `workflow_dispatch`, or `just deploy-stg <ref>`                  | None                                                     |
| Cloud Build trigger | None. Developer Connect repositories cannot have manual triggers | Yes, defined in Terraform                                |
| Approval            | Not required                                                     | Required. Pushing the tag alone does not start the build |
| Image tag           | The commit SHA                                                   | The commit SHA                                           |

- Terraform ignores the `image` of the Cloud Run service with `ignore_changes`, so a deployment is not reported as drift.
- All GitHub Actions does is start the Cloud Build. `sa-deployer` still performs the build and the deployment inside GCP, so no deployment permission is handed to GitHub.
- The steps on this page are for starting a deployment by hand. Anything merged into `main` is deployed by `cd-backend-stg`, so running them is normally unnecessary.
- The steps for creating the Developer Connect connection live in the design notes, in the section titled "5. Developer Connect の接続" (Setting up the Developer Connect connection).
