# cronlog

Structured log aggregator for cron jobs with failure notifications and retention policies.

## Installation

```bash
go install github.com/yourusername/cronlog/cmd/cronlog@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/cronlog.git && cd cronlog && make install
```

## Usage

Wrap any cron command with `cronlog` to capture structured output, detect failures, and apply retention rules.

```bash
# Basic usage
cronlog run --job "backup" -- /usr/local/bin/backup.sh

# With failure notification and log retention
cronlog run \
  --job "db-cleanup" \
  --notify slack \
  --retain 30d \
  -- /usr/local/bin/db-cleanup.sh
```

Configure jobs via `cronlog.yaml`:

```yaml
jobs:
  db-cleanup:
    command: /usr/local/bin/db-cleanup.sh
    retain: 30d
    notify:
      on_failure: true
      channel: slack
      webhook: https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

View aggregated logs:

```bash
# List recent job runs
cronlog logs --job "db-cleanup" --last 10

# Show details for a specific run
cronlog logs --job "db-cleanup" --run-id abc123
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `--job` | required | Job name identifier |
| `--retain` | `7d` | Log retention duration |
| `--notify` | none | Notification provider (`slack`, `email`) |
| `--config` | `cronlog.yaml` | Path to config file |

## License

MIT — see [LICENSE](LICENSE) for details.