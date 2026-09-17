# Connecting to the stg Firestore

[English](./stg.md) | [日本語](./stg.ja.md)

Back to the [documentation index](../../README.md).

This runs the local api against the Firestore in the GCP project `stg-aozora-park` instead of the emulator.
The `(default)` database is already created by Terraform (`infra/envs/stg`).

> [!WARNING]
> This writes to a real database. Delete any documents you create for testing once you are done.

## Prerequisites

- You can read and write Firestore in `stg-aozora-park` (for example, as a project owner)
- Application Default Credentials (ADC) are set up

```sh
gcloud auth application-default print-access-token > /dev/null && echo ADC OK
```

If `ADC OK` is not printed, log in.

```sh
gcloud auth application-default login
```

## Running

```sh
cd backend
just run-api-stg
```

How it differs from `just run-api`:

| Item                         | `just run-api`     | `just run-api-stg` |
| ---------------------------- | ------------------ | ------------------ |
| Target                       | Emulator           | stg Firestore      |
| `GOOGLE_CLOUD_PROJECT`       | `demo-aozora-park` | `stg-aozora-park`  |
| `FIRESTORE_EMULATOR_HOST`    | `localhost:18080`  | Not set            |
| `GOOGLE_CLOUD_QUOTA_PROJECT` | Not set            | `stg-aozora-park`  |

- If `FIRESTORE_EMULATOR_HOST` is left in your shell, the SDK connects to the emulator. The recipe unsets it before starting.
- The quota project used for API billing is pinned to `stg-aozora-park` rather than the value recorded in ADC.
- The port and the allowed origin (`http://localhost:5173`) are the same as `just run-api`, so the frontend's `just dev` can call it as is.

## Checking

With the api running, create, get and update a park from the UI or with [`backend/tools/http/park.http`](../../../backend/tools/http/park.http).
Open the `.http` file with the VS Code extension REST Client (`humao.rest-client`) and run each request with the "Send Request" link above it.
Get and update target the park returned by the last create, so run create first.

The documents appear in the `parks` collection in [Firestore Studio](https://console.cloud.google.com/firestore/databases/-default-/data/panel?project=stg-aozora-park).
