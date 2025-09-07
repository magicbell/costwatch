# CostWatch

CostWatch pulls usage metrics (e.g., from AWS CloudWatch), stores them in ClickHouse, and exposes an API and dashboard for cost analysis and projections.

## Quick start (Docker)

Prerequisite: Docker Desktop (or Docker Engine) installed and running.

1) Clone the repo and change into the project folder:

```shell
git clone https://github.com/costwatchai/costwatch
cd costwatch
```

2) Start the full stack:

```shell
docker compose up
```

3) Visit the dashboard at http://localhost:3000

Press Ctrl+C to stop. To run in the background: `docker compose up -d` and stop with `docker compose down`.

## Configuration

Most users can run with the defaults. If needed, copy example.env to .env and adjust values:

```
cp example.env .env
```

## Contributing / local development

See CONTRIBUTING.md for a workflow that runs services locally (without Docker) and provides convenient tasks for development.

## Troubleshooting

- Ports 3000, 3010, 3020, 9000, or 8123 already in use:
  - Stop other processes using these ports or change the mappings in docker-compose.yml
- First run takes a while:
  - Docker may need to pull base images; subsequent runs will be faster
- Dashboard can't reach API from Docker:
  - The dashboard container generates its client pointing to http://localhost:3010/v1 and should just work via compose networking; ensure the API is up and healthy
