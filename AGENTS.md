# Server development guide

This monorepo runs the local API, internal MCP server, LangChain agent, and two workers. Start from the service's normal entry point and trace its shared dependencies and downstream consumers. Verify the change along that path.

## Layout and responsibilities

- Keep Go service entry points under `cmd/{api,mcp,image-extract-worker,index-worker}`. Domain logic belongs in the corresponding `internal` package. Shared storage, model access, and observability integrations live under `internal/platform`.
- TypeScript agent code and prompts live in `agent/`, versioned contracts in `contracts/`, database changes in `migrations/`, and local deployment files in `deploy/` and the root Compose file. Inspect existing callers before extending a package.
- The API owns browser context, rights, and credit accounting. The agent invokes API operations through internal MCP. Extraction produces validated analysis JSON. Indexing writes both vectors on the same point and completes publication. Carry object keys and versions in queue messages.

## Compatibility and verification

HTTP fields and status codes, WebSocket events, MCP tools, and RabbitMQ messages are contracts with independent consumers. Make additive fields compatible with older inputs. For a breaking change, deploy readers that understand both versions before switching writers. Test an old queued message and an earlier client request. Verify migrations against existing rows. For an embedding model or dimension change, populate and validate a new collection before switching the search alias.

When changing credit accounting, generation, publication, or acknowledgement order, replay a request and simulate interruption after a write. Check the file, balance, point, and reward for a single consistent outcome.

## Observability

Preserve request and operation IDs across services. If latency, errors, provider costs, or queue behavior change, exercise the Prometheus metric and its Grafana panel together. Review existing queries before changing metric names, units, or labels. Keep request and asset IDs in logs rather than metric labels. When Sentry is enabled, include service and release context for new errors. Local Compose must work with an empty DSN. Exclude raw audio, private images, cookies, and provider credentials from telemetry.

## Delivery and release notes

From the repository root run `go test ./...` and `docker compose config --quiet`. Rebuild the affected service and exercise it from its real consumer. For telemetry changes, check the target in `deploy/prometheus.yml` and the associated query in `deploy/grafana/dashboards/`. Redelivering an indexing job may increase the attempt count. Verify a single publication and contribution reward.

Record user-visible, integration, migration, and operational changes under **Unreleased** in [CHANGELOG.md](CHANGELOG.md) in the same change set. Describe who is affected, the observable change, and migration steps. Cross-repository changes require an entry for each affected repository. Group delivered entries under the actual release version and date.
