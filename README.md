# CostWatch

Cloud providers like AWS offer billing metrics and cost alerts, but this data is typically delayed by 6-24 hours. During this delay window, unexpected usage spikes can go unnoticed, potentially resulting in significant cost overruns that are only discovered after the fact.

Costwatch solves this problem by leveraging near-real-time usage metrics (like CloudWatch metrics) and associating them with their corresponding costs. This enables immediate detection of cost anomalies and real-time alerting on spending spikes, allowing teams to respond to issues before they become expensive surprises.

CostWatch pulls usage metrics (e.g., from AWS CloudWatch), stores them in ClickHouse, and exposes an API and dashboard for cost analysis, projections, and alerts. It is developed at, and in use at [MagicBell](https://www.magicbell.com). 

![screenshot.png](docs/screenshot.png)

## Quick start (Docker)

Prerequisite: Docker Desktop (or Docker Engine) installed and running.

1. Clone the repo and change into the project folder:

```shell
git clone https://github.com/magicbell/costwatch
cd costwatch
```

Copy the `example.env` file to `.env`

```shell
cp example.env .env
```

2. Start the full stack:

```shell
docker compose up
```

3. Visit the dashboard at http://localhost:3000

Press Ctrl+C to stop. To run in the background: `docker compose up -d` and stop with `docker compose down`.

### Demo data (CoinGecko)

For demo purposes, CostWatch ships with a default provider that fetches BTC price metrics from CoinGecko. This helps populate charts immediately after startup.

- Enabled by default when you follow the quick start.
- You can disable this demo provider by setting the environment variable `DEMO` to `off` or `false`.

```shell
DEMO=off docker compose up
```

## AWS authentication

CostWatch talks to AWS (e.g., CloudWatch) through the standard AWS SDK credential chain. Ensure your AWS credentials are available in the environment where the API/worker run.

Common ways this works:

- AWS SSO: log in with `aws sso login` for your profile; the SDK will pick up your session from your AWS config/credentials files.
- Shared credentials/config files: `~/.aws/credentials` and `~/.aws/config`.
- Environment variables: `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, and, when applicable, `AWS_SESSION_TOKEN`.

Verify your identity with the AWS CLI:

```bash
aws sts get-caller-identity
```

If you are running via Docker, make sure the containers can access your credentials (for example, by exporting the env vars before `docker compose up`, or by mounting your `~/.aws` directory if that matches your workflow).

## Receiving alerts

Alerts are optional and can be posted to a Slack‑compatible incoming webhook.

- Set `ALERT_WEBHOOK_URL` in your `.env` to your webhook URL (for example, a Slack Incoming Webhook). The notifier posts a simple JSON payload like `{ "text": "..." }`, which is also compatible with many Slack‑compatible systems (e.g., Mattermost, Rocket.Chat).
- Bring the stack up with `docker compose up` (compose loads `.env` for the api/worker).
- Configure alert rules either:
  - In the dashboard (Hourly costs card, Alert threshold column) when using SQLite; or
  - Via environment variable `ALERT_RULES` for read‑only environments. Example:
    `ALERT_RULES='[{"service":"aws.CloudWatch","metric":"IncomingBytes","threshold":0.47}]'`
- When thresholds are exceeded, the worker will send notifications to ALERT_WEBHOOK_URL.
- For ongoing incidents, alerts will be sent at most once an hour.

Notes:
- When `ALERT_RULES` is set, alert rules are read‑only and persisted changes via the API are disabled.
- In env mode, last notification timestamps are not recorded; the system behaves as if never notified before.

Tip: you can copy the provided example and then edit it:

```shell
cp example.env .env
# open .env and set ALERT_WEBHOOK_URL=https://hooks.slack.com/services/...
```

## Running on Lambda

The API & Dashboard server support Lambda Function URL invocation signature. It is automatically enabled in the Lambda runtime where the environment variable `AWS_LAMBDA_FUNCTION_NAME` is set.

## Contributing / local development

See [CONTRIBUTING.md](/CONTRIBUTING.md) for a workflow that runs services locally (without Docker) and provides convenient tasks for development.

## Troubleshooting

- Ports 3000, 3010, 3020, 9000, or 8123 already in use:
  - Stop other processes using these ports or change the mappings in docker-compose.yml
- The first run takes a while:
  - Docker may need to pull base images; later runs will be faster
- Dashboard can't reach API from Docker:
  - The dashboard container generates its client pointing to http://localhost:3010/v1 and should just work via compose networking; ensure the API is up and healthy

```

```
