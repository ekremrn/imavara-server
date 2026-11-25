# Server changelog

Follow the release and entry guidelines in [AGENTS.md](AGENTS.md).

## Unreleased

- Published the repository guides and architecture documentation in English.
- Format agent responses in Markdown, separating text across tool calls into paragraphs.
- Enabled the RabbitMQ management portal on the Pi at localhost port 15672.
- Added optional `top_k` to API and MCP catalog searches. The limit is 1–10, with a default of 5.
- Return similarity scores and an optional nearest candidate when there is no strong match.
- Automatically submit delivered generations and variations for catalog admission, including backfilling previously private generations.
- Run the local catalog on one Qdrant node with 16 shards and replication factor 1. Compose uses the new `qdrant_data` volume. See the README for migration instructions.
- Added development and changelog guides, service verification commands, and metric checks for redelivered jobs.
- Added the local Go API and PostgreSQL setup, anonymous browser sessions, initial grants, pricing, and a credit ledger.
- Added chat history, voice transcription, internal MCP tools, and the LangChain agent.
- Added catalog search, file access, acquisition, generation, contributions, queue processing, and service dashboards. Live model acceptance checks remain in progress.
- Added WebP originals, actual and requested aspect ratios, private reference uploads, a Turkish transcription language hint, and a failed-job replay command.
