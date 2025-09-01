# CostWatch

CostWatch is a service that pulls usage metrics (e.g., from AWS CloudWatch), stores them in ClickHouse, and exposes a dashboard server for cost analysis and projections.

This repository is designed to be run with [Task](https://taskfile.dev/), [process-compose](https://f1bonacc1.github.io/process-compose), and [Docker](http://docker.com). The single command `task up` will:

- install local dependencies (via the Taskfile)
- start Docker services (ClickHouse)
- seed the ClickHouse schema
- run the development server(s)

## Prerequisites

Install the following tools before you start:

- Docker https://www.docker.com/products/docker-desktop/
- Task https://taskfile.dev
- Process Compose https://f1bonacc1.github.io/process-compose/

## Quick start

1) Clone the repo and change into the project folder:

```
git clone https://github.com/costwatchai/costwatch
cd costwatch
```

2) Start everything with a single command:

```
task up
```

## Common tasks

- Start everything (preferred)
  - `task up`

- Seed ClickHouse schema manually (if needed)
  - `task seed-clickhouse`

- Run the Go API locally without the full stack (requires ClickHouse to be available)
  - `task run-api`

- Lint Go code
  - `task lint`

- Stop Docker services started by process-compose
  - From another terminal: `docker compose stop` (in the repo root)
  - Or stop via your Docker Desktop UI


## Environment variables

Task reads `.env` if present (see `dotenv: ['.env']` in taskfile.yml). You can create a `.env` at the repo root to override defaults.


## Troubleshooting

- process-compose not found:
  - Install with Homebrew (`brew install process-compose`) or follow docs: https://f1bonacc1.github.io/process-compose/
- task not found:
  - Install with Homebrew (`brew install go-task/tap/go-task`) or with Go (`go install github.com/go-task/task/v3/cmd/task@latest`)
- Docker not running:
  - Start Docker Desktop (macOS/Windows) or ensure the Docker daemon is running (Linux)
- Ports 9000 or 8123 already in use:
  - Stop other ClickHouse/containers using these ports or change the mappings in docker-compose.yml
- Seeding ClickHouse fails with connection errors on first run:
  - Wait for ClickHouse to become healthy; process-compose’s seed step depends on ClickHouse logs but your machine may need extra time on initial image pull
- Go build issues:
  - Ensure you have a recent Go version installed, run `go version`, and try `go mod download`


## Development notes

- The seed command is implemented in `cmd/admin/admin.go` and intentionally connects to the ClickHouse `default` database to create the target `costwatch` database and its `metrics` table.
- The application connects to the `costwatch` database by default (see `cmd/main.go`).
- The ClickHouse schema is defined in `internal/clickstore/schema.go`.

Feel free to open issues or PRs for enhancements to the Taskfile and process-compose flow.
