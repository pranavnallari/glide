# Glide

A lightweight LLM gateway written in Go. Glide sits in front of multiple LLM providers (OpenAI, Anthropic) and exposes a single unified API — handling routing, fallback, rate limiting, and observability so your application doesn't have to.

## Why Glide?

Calling LLM APIs directly couples your application to a single provider. Glide decouples this:

- **Provider resilience** — if OpenAI is down, fall back to Anthropic automatically
- **Cost control** — route cheap requests to smaller models, expensive ones to powerful ones
- **Rate limiting** — token bucket limiting at global, per-provider, and per-key levels
- **Observability** — Prometheus metrics for request counts, latency, and token usage out of the box

## Architecture

```
HTTP Client
    │
    ▼
POST /chat
    │
    ▼
┌─────────────────────────────┐
│         API Handler         │
│  global + per-key limiting  │
│      metrics recording      │
└─────────────┬───────────────┘
              │
              ▼
┌─────────────────────────────┐
│           Router            │
│  fallback / priority /      │
│  load_balance strategies    │
│  per-provider limiting      │
└──────┬──────────────┬───────┘
       │              │
       ▼              ▼
  OpenAI          Anthropic
  Adapter         Adapter
       │              │
       ▼              ▼
  api.openai.com  api.anthropic.com
```

## Routing strategies

Glide supports three routing strategies, configured per route in `config.yaml`:

**fallback** — tries providers in order, moving to the next on any error. Best for reliability.
```yaml
default:
  strategy: fallback
  order:
    - provider: openai
      model: gpt-4o
    - provider: anthropic
      model: claude-sonnet-4-6
```

**priority** — like fallback but skips disabled providers. Good for gradual provider migration.
```yaml
cheap:
  strategy: priority
  order:
    - provider: openai
      model: gpt-4o-mini
```

**load_balance** — round-robin across a pool of providers. Good for spreading load.
```yaml
fast:
  strategy: load_balance
  pool:
    - provider: openai
      model: gpt-4o-mini
    - provider: anthropic
      model: claude-sonnet-4-6
```

## Rate limiting

Glide uses a token bucket algorithm with three independent levels:

- **Global** — caps total requests across all providers
- **Per-provider** — caps requests to each provider independently
- **Per-key** — caps requests per API key, for multi-tenant setups

All levels are opt-in — omitting a limit in config means unlimited. A request must pass all configured levels to proceed; hitting any limit returns `429 Too Many Requests`.

## Getting started

**Prerequisites**: Go 1.22+, Docker (optional)

```bash
git clone https://github.com/pranavnallari/llm-gateway
cd llm-gateway
```

Set your API keys:
```bash
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
```

Run:
```bash
make run
```

Or with the full observability stack (Glide + Prometheus + Grafana):
```bash
make docker-up
```

## Sending a request

```bash
curl -X POST http://localhost:8080/chat \
  -H "Content-Type: application/json" \
  -H "Authorization: my-app-key" \
  -d '{
    "messages": [{"role": "user", "content": "Hello!"}],
    "strategy": "default"
  }'
```

Response:
```json
{
  "id": "chatcmpl-abc123",
  "provider": "openai",
  "content": "Hello! How can I help you today?",
  "model": "gpt-4o",
  "usage": {
    "prompt_tokens": 10,
    "completion_tokens": 9,
    "total_tokens": 19
  }
}
```

The `strategy` field selects which route from `config.yaml` to use. Defaults to `"default"` if omitted or unrecognised.

## Health check

```bash
curl http://localhost:8080/health
```

Returns `200 OK` when the server is running.

## Configuration

Full `config.yaml` reference:

```yaml
server:
  port: 8080
  timeout: 30s

providers:
  openai:
    enabled: true
    api_key: ${OPENAI_API_KEY}       # resolved from environment
    base_url: https://api.openai.com/v1
    models:
      - name: gpt-4o
        input_cost: 0.005            # USD per 1k tokens
        output_cost: 0.015
        rpm: 500
        tpm: 100000

  anthropic:
    enabled: true
    api_key: ${ANTHROPIC_API_KEY}
    base_url: https://api.anthropic.com
    models:
      - name: claude-sonnet-4-6
        input_cost: 0.003
        output_cost: 0.015
        rpm: 200
        tpm: 80000

routing:
  default:
    strategy: fallback
    order:
      - provider: openai
        model: gpt-4o
      - provider: anthropic
        model: claude-sonnet-4-6

limits:
  global:
    rpm: 2000
  per_provider:
    openai:
      rpm: 1000
    anthropic:
      rpm: 500

keys:
  my-app-key:
    rpm: 100

tracking:
  enabled: true
  currency: USD
  log_usage: true
```

## Metrics

Glide exposes a Prometheus `/metrics` endpoint. Key metrics:

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `glide_requests_total` | Counter | `provider`, `status` | Total routed requests |
| `glide_request_duration_seconds` | Histogram | `provider` | Provider call latency |
| `glide_tokens_total` | Counter | `provider`, `token_type` | Token usage (input/output) |
| `glide_cost_total` | Counter | `provider` | Estimated cost in USD |

```bash
curl http://localhost:8080/metrics
```

With `make docker-up`, Grafana is available at `http://localhost:3000` (admin/admin). Add Prometheus (`http://prometheus:9090`) as a data source to start building dashboards.

## Running tests

```bash
make test
```

## Make targets

| Target | Description |
|--------|-------------|
| `make run` | Run the app |
| `make build` | Compile binary to `bin/glide` |
| `make test` | Run all tests |
| `make test-verbose` | Run tests with full output |
| `make lint` | Run `go vet` |
| `make clean` | Remove `bin/` |
| `make docker-up` | Build and start full stack |
| `make docker-down` | Tear down Docker stack |


## Project structure

```
.
├── cmd/glide/          # entry point
├── internal/
│   ├── adapter/        # provider-specific translation (OpenAI, Anthropic)
│   ├── api/            # HTTP handler and routing
│   ├── config/         # config loading and env var resolution
│   ├── limiter/        # token bucket rate limiter
│   ├── models/         # unified request/response types
│   ├── observability/  # Prometheus metrics
│   └── router/         # routing strategies
├── config.yaml         # example configuration
├── prometheus.yml      # Prometheus scrape config
└── docker-compose.yml  # full observability stack
```