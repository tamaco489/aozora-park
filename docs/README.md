# Documentation index

[English](./README.md) | [日本語](./README.ja.md)

This directory contains project documentation.

| Document                                                          | What it covers                                                     |
| ----------------------------------------------------------------- | ------------------------------------------------------------------ |
| [Backend package layout](./backend/packages/overview.md)          | Layers, dependencies and the decisions behind them                 |
| [Firestore emulator](./backend/firestore/emulator.md)             | Running the emulator locally and connecting to it                  |
| [Connecting to the stg Firestore](./backend/firestore/stg.md)     | Pointing the local api at the stg Firestore                        |
| [API specification (OpenAPI)](./api/openapi.yaml)                 | Generated from proto; do not edit by hand                          |
| [API specification (Redoc)](./api/redoc.html)                     | HTML version of the above; open it in a browser                    |
| [Deployment architecture](./cd/overview.md)                       | An overview comparing the backend and frontend paths               |
| [Workload Identity Federation](./cd/wif/overview.md)              | How GitHub Actions reaches Google Cloud without storing a key      |
| [Backend deployment architecture](./backend/deploy/overview.md)   | Cloud Build and Developer Connect, and what Terraform owns         |
| [Deploying the backend to stg](./backend/deploy/stg.md)           | Updating the stg api from your machine, verification and rollback  |
| [Frontend deployment architecture](./frontend/deploy/overview.md) | Same-origin rewrites and caching, and the alternatives rejected    |
| [Deploying the frontend to stg](./frontend/deploy/stg.md)         | Updating the stg site from your machine, verification and rollback |
