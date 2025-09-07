# Contributing

This repository uses [mise](https://mise.jdx.dev) to provide a simple, reproducible developer experience without requiring Docker.

## Prerequisites

- Install mise: https://mise.jdx.dev/getting-started.html#installing-mise-cli

## Quick start (local dev)

From the repo root:

```shell
mise dev
```

## Services & Ports
  
App | Start cmd | PORT
--- | --- | ---
Dashboard | `mise run-dashboard` | 3000
API | `mise run-api` | 3010
Worker | `mise run-worker` | 3020

## Regen Dashboard API Client

The API client & types for the dashboard, are generated using the openapi spec from the API. Run the command below to regenerate the client & types.

```shell
mise gen-api-client
```

## Lint code

Run the following command to lint both go and typescript code:

```
mise lint
```

## Alternative: run everything with Docker

The easiest way to get started is to simply run:

```
docker compose up
```

This will start ClickHouse, seed the schema, and run API, worker, and dashboard.
