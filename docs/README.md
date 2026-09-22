# aozora-park

[English](./README.md) | [日本語](./README.ja.md)

This directory contains project documentation.

| Document                                                      | What it covers                                                    |
| ------------------------------------------------------------- | ----------------------------------------------------------------- |
| [Backend package layout](./backend/packages/overview.md)      | Layers, dependencies and the decisions behind them                |
| [Firestore emulator](./backend/firestore/emulator.md)         | Running the emulator locally and connecting to it                 |
| [Connecting to the stg Firestore](./backend/firestore/stg.md) | Pointing the local api at the stg Firestore                       |
| [API specification (OpenAPI)](./api/openapi.yaml)             | Generated from proto; do not edit by hand                         |
| [API specification (Redoc)](./api/redoc.html)                 | HTML version of the above; open it in a browser                   |
| [Deployment architecture](./deploy/overview.md)               | The paths, stg versus prd, and what Terraform owns                |
| [Deploying to stg](./deploy/stg.md)                           | Updating the stg api from your machine, verification and rollback |
