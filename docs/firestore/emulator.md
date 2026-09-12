# Firestore emulator

[English](./emulator.md) | [日本語](./emulator.ja.md)

Back to the [documentation index](../README.md).

Local development uses the Firestore emulator from the Firebase Emulator Suite.
It runs in a container, so neither a JDK nor the Firebase CLI is needed on the host.

## Start and stop

```sh
just up
just down
```

`just up` waits until the emulator reports healthy. The image is built on the first run. It takes a few minutes because it installs a JRE and the Firebase CLI.

## Ports

| Port    | What it serves                                             |
| ------- | ---------------------------------------------------------- |
| `18080` | Firestore emulator. `8080` onwards is left for api servers |
| `4000`  | Emulator UI                                                |
| `4400`  | Emulator hub                                               |
| `4500`  | Log stream the UI reads                                    |
| `9150`  | Websocket the UI uses to receive changes                   |

Open the UI at <http://localhost:4000/firestore>.

## Connecting a client

Set the host and the SDK talks to the emulator instead of Google Cloud.

```sh
export FIRESTORE_EMULATOR_HOST=localhost:18080
```

The local project ID is `demo-aozora-park`.
An ID starting with `demo-` makes the SDK refuse to reach the real Google Cloud,
so a misconfigured client cannot touch production data by accident.

## Connecting the api

`backend/justfile` passes both variables to the recipes that need them, so no extra setup is required.

```sh
cd backend
just run-api  # starts the api against the emulator
```

`GOOGLE_CLOUD_PROJECT` has no default in the code. The api refuses to start without it,
so a misconfigured deployment fails at startup instead of talking to the wrong project.

## Tests use their own emulator

This container is for running the app by hand. Tests do not use it.

```sh
cd backend
just test           # skips the tests that need an emulator
just test-emulator  # runs them too, starting a container per package
```

Tests start their own emulator with testcontainers and gate it on `DOCKER_TESTS`.
Without the variable they skip themselves, so `just test` stays runnable where Docker is missing.
Keeping the two apart means a test never reads or deletes the data you are looking at in the UI.

## Verifying read and write

```sh
BASE="http://localhost:18080/v1/projects/demo-aozora-park/databases/(default)/documents"

curl -X POST "$BASE/parks?documentId=aozora" \
  -H 'Content-Type: application/json' \
  -d '{"fields":{"name":{"stringValue":"Aozora Park"}}}'

curl "$BASE/parks/aozora"
```

Data lives in memory only. Everything is gone once the container stops.
