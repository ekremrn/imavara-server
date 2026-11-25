# MVP service tasks and implementation guide

The server runs five backend processes: Go API, Go MCP, TypeScript web agent, extraction worker, and index worker. React runs in its own repository. Directory names below indicate code ownership. Unchecked boxes track pending verification. K numbers refer to the [delivery plan](DELIVERY-PLAN.md). Use the [product plan](IMAVARA-MVP.md) for endpoints, model profiles, and filter types.

## Implementation order

| Step | Tasks | Checkpoint |
| --- | --- | --- |
| 1. Prepare the environment | ORT-01–05 | Compose starts on empty volumes, the collection is ready, and all seven model routes accept real samples. |
| 2. Establish sessions and file rights | API-01–03 | Verify balances, ownership, and repeat downloads from two browsers. |
| 3. Find the first contribution | API-04, EXT-01–03, IDX-01–04, API-05 | Upload a licensed photo, find it through filtered search from a second browser, and submit it again to check the reward. |
| 4. Connect the catalog to the web | API-06, WEB-01–02 | Exercise Discover, acquisition, batch upload, and popularity. |
| 5. Complete generation | API-07, WEB-03 | Deliver a new image or variation of a selected card in one HTTP request. |
| 6. Connect chat and voice | API-08–09, MCP-01–03, AGT-01–03, WEB-04–05 | Describe the same need by text and voice. Observe real tool calls. |
| 7. Verify delivery | SON-01–03 | Measure quality, costs, failures, and the final demo on one version. |

Within each section, work in order and verify the output through its downstream consumer. For each completed K item, record the command or user action, outcome, and relevant request or asset ID.

Follow the affected repository's AGENTS.md before closing a task. When adding a card field, verify that old clients read new messages and that the new client opens a conversation missing that field. Exercise relevant Prometheus queries, Grafana panels, and enabled Sentry error context when counts or failure behavior change. Record changes affecting users, other services, or operators under Unreleased. [Server changelog](../CHANGELOG.md) · [Web changelog](https://github.com/ekremrn/imavara-web/blob/main/CHANGELOG.md)

## Shared infrastructure and contracts

Target directories include `contracts`, `internal/platform`, `migrations`, and `deploy`.

- [ ] **ORT-01 · Repositories and network.** Prepare both repos following the [web guide](https://github.com/ekremrn/imavara-web/blob/main/AGENTS.md) and [server guide](../AGENTS.md). Use one Go module, one TypeScript agent package, and separate container entry points for the five processes. Add PostgreSQL, Qdrant, RabbitMQ, Garage, Prometheus, and Grafana to Compose. Define startup, example environment, health checks, and volumes. Route users through web/API. Keep OpenRouter credentials and Sentry DSN server-side. Verify startup with an empty DSN.
- [ ] **ORT-02 · Contracts.** Version cards, errors, MCP inputs/results, WebSocket events, and both queue messages. Specify which request, operation, message, and asset IDs survive retries. Carry asset ID, object key, and version in queue messages. Verify web and agent consume the same card and event ordering.
- [ ] **ORT-03 · Data services.** Prepare migrations and uniqueness rules. Isolate original, preview, and private analysis access in Garage. Provision extraction, index, and failed queues. Use 16 shards, one node, replication factor 1, two named cosine vectors per point, binary quantization, and rescoring. Create the payload indexes in section 11 before the first upload. Rerun bootstrap and verify data survives.
- [ ] **ORT-04 · Clients and telemetry.** Share Go storage, OpenRouter, queue clients, configuration, structured logging, and correlation IDs. Scrape Prometheus metrics from every service and correlate TypeScript agent logs. Share embedding length validation and normalization between ingestion and queries.
- [ ] **ORT-05 · Provider checks.** Exercise all seven fixed profiles. Verify Ember tool calls, Muse generation and edits, Luna decisions, valid Glimmer JSON, Qwen and Gemini dimensions, and Turkish transcription. Record outcomes, cost, and incompatibilities by model ID. **K1, K2, K11**

## Go API

Entry point: `cmd/api`. Accept web HTTP/WebSocket requests and operations forwarded by MCP. Own session context, access decisions, credits, and file delivery.

- [ ] **API-01 · Browser session.** Bind a random HttpOnly cookie to a PostgreSQL session. Create the session and +100 grant atomically. Expose profile, pricing, and history. Check two tabs of one browser and an independent second browser. **K7, K8**
- [ ] **API-02 · Credits.** Reserve, settle, release, and deduplicate first acquisitions. Replay a paid request and return its prior result. Race two requests for the last three credits. Define the unique publication reward in the same ledger. **K7**
- [ ] **API-03 · Files and rights.** Read actual file format, dimensions, and SHA-256. Persist asset, preview, and original keys. Enforce browser rights for details, library, acquisition, stable downloads, sharing, and removal. First acquisition costs 1 and repeat download costs zero. Verify existing entitlements survive normal withdrawal while new acquisitions close. **K5, K7, K8**
- [ ] **API-04 · Contribution admission.** Record per-file permission and terms. Check content-derived Qdrant point ID and full hash, including concurrent submissions. Distinguish near matches from exact copies. Submit candidates to Luna and distinguish rejection, review, provider error, and acceptance. Persist admission and pending work durably across broker outages. **K9, K10**
- [ ] **API-05 · Search.** Use Qwen for text and Gemini or a stored accessible vector for reference images. Apply the appropriate named vector, visibility, license, strict user filters, and field threshold. Default to five cards and accept `top_k` from 1–10. Check the ledger after successful and empty searches. **K3, K11, K12**
- [ ] **API-06 · Gallery and statistics.** Paginate recent uploads. Record each returned asset and rank once per search. Calculate period-specific rankings from first acquisitions and search events separately. Two searches, one acquisition, and two downloads must produce two returns and one acquisition. **K5, K13**
- [ ] **API-07 · Generation and variations.** On explicit request reserve 3 credits. Pass the authorized reference, edit instructions, and ratio to Muse Image. Store the original in Garage, record actual dimensions and ownership, settle delivery, and create the catalog job transactionally. Check definite/transient failures, interrupted responses, and repeated operation keys. **K6, K7, K8**
- [ ] **API-08 · Chat.** Create, list, open, and delete conversations. Resolve the browser server-side, pass messages and accessible references to the agent, and save messages, tool results, and card order. Reconcile state on WebSocket reconnect. One message ID starts one agent turn. **K3, K5, K8**
- [ ] **API-09 · Voice.** Validate format and size at `POST /v1/audio/transcriptions`. Return text from `openai/gpt-transcribe`. Retain audio only during processing. Distinguish empty text and provider errors from a valid message. **K2, K4**

## Extraction worker

Entry point: `cmd/image-extract-worker`. Consume a RabbitMQ job and the Garage original. Produce a validated analysis file and an index job.

- [ ] **EXT-01 · Receive work.** Validate the message version and reuse completed analysis for the same asset version. Otherwise read the object and invoke `meta/muse-glimmer-30b` at fixed low effort.
- [ ] **EXT-02 · Validate analysis.** Apply JSON Schema and cross-field rules. Test unreadable lettering, unknown attributes, and copy-space direction. Store the valid result privately with asset, prompt, model, and schema versions. [Analysis schema](../contracts/glimmer-analysis.v1.schema.json)
- [ ] **EXT-03 · Hand off indexing.** Publish the index message and ACK extraction after broker confirmation. Bound transient retries and route persistently invalid results to the failure queue. Restart after writing but before ACK and verify the correct version reaches indexing. **K2, K10, K13**

## Index worker

Entry point: `cmd/index-worker`. Read validated analysis, original, and current sharing state. Produce a searchable point, publication record, and eligible contribution reward.

- [ ] **IDX-01 · Search document and filters.** Combine description, main subject, visible relationships, composition, and tags deterministically. Map style, color, and copy space to payload. Get dimensions, format, ratio, and hash from the file record. Omit unknown boolean fields.
- [ ] **IDX-02 · Embeddings.** Process text with `qwen/qwen3-embedding-8b` and pixels with `google/gemini-embedding-2`. Share transformations with search. Fail the job on invalid length or non-finite elements.
- [ ] **IDX-03 · Publish.** Check current sharing version, write both vectors and filters with the stable point ID, and prevent old work from republishing withdrawn assets. Deduplicate publication and external contribution reward. ACK after durable completion.
- [ ] **IDX-04 · Interruption tests.** Interrupt after Qdrant writes and before reward, then after reward and before ACK. Redeliver and check one visible asset, one reward, and both search modes. Withdraw during processing and search from a second browser. **K7, K8, K9, K10, K11**

## Go MCP

Entry point: `cmd/mcp`. Receive internal agent tool calls and return API outcomes in the same browser and operation context.

- [ ] **MCP-01 · Register tools.** Define input and output contracts for `get_balance`, `search_images`, `get_image`, `acquire_image`, `generate_image`, and `contribute_image`. Inspect the list and schemas from the agent connection.
- [ ] **MCP-02 · Forward to the API.** Use trusted session context. Prevent tools from selecting a browser via arguments. Preserve reference IDs, filters, paid-operation keys, card order, actual charges, and errors.
- [ ] **MCP-03 · Compare entry points.** Compare filtered API and MCP results in the same session. Replay a paid key and verify one movement. Check private asset isolation with a second browser and internal-only MCP reachability. **K1, K3, K7, K8**

## TypeScript web agent

Directory: `agent/`. Receive conversation and reference context from the API, invoke internal MCP tools through LangChain, and stream events to the API.

- [ ] **AGT-01 · Model and tools.** Load the [Turkish system prompt](../agent/prompts/system.tr.txt). Connect MCP and `meta/muse-ember` with temperature 0.3, low effort, automatic tool choice, and streaming.
- [ ] **AGT-02 · Tool decisions.** Test one clarifying question for an ambiguous brief, free search for explicit lookup, and paid generation for an explicit request. “Find something like the second one” and “remove the cup from the second one” use the same asset with different tools. Wait for the user's decision before generating after an empty result.
- [ ] **AGT-03 · Stream events.** Forward actual tool starts, cards, text chunks, delivery, and errors with message ID and order. Compare the displayed charge with API records after replay and tool failure. **K2, K3, K5, K6, K7**

## React web application

Repository: `imavara-web`. The browser uses the product API and chat WebSocket.

- [ ] **WEB-01 · Sessions and navigation.** Connect session, balance, pricing, and errors through the shared client. Build the minimal landing field, Meta Muse credit, Discover, and Library.
- [ ] **WEB-02 · Discovery and contributions.** Wire text/reference search, filters, original framing, pagination, recent and popular assets, and acquisition. For batch uploads, display outcomes per file and updated balance. Test an accepted and rejected file in one selection.
- [ ] **WEB-03 · Generation and library.** Generate from scratch and vary a card. Display actual file dimensions. Support delivery, automatic catalog review, withdrawal, and repeat download of acquired originals.
- [ ] **WEB-04 · Chat.** Transition from landing on the first message. Preserve card order in WebSocket events. Support new chats, history, and reconnect. Check the acquired asset remains in Library after deleting the conversation.
- [ ] **WEB-05 · Voice.** Add microphone input to the landing page and chat. Show duration, Send, Cancel, and waiting status. Auto-submit one message combining transcript, draft, and reference. Exercise double tap, permission denial, empty recording, and network failure. Verify drafts and message counts before closing K4. **K3, K4, K5, K6, K8, K13**

## Cross-service verification

- [ ] **SON-01 · Search and model measurement.** Ingest licensed seed data through the contribution pipeline. Split about 100 labeled Turkish/English requests and additional reference images into calibration and acceptance sets. Tune distinct text and image thresholds. Compare held-out top-five hits and correct empty results against targets. Record exact-search comparison, p95 latency, and cost per valid Glimmer analysis for the first 100 representative images. **K2, K12**
- [ ] **SON-02 · Operations.** Inspect all five Grafana dashboards. Trace an upload from API through workers. Exercise model faults, queue age, time to publication, cost, and empty searches. Restore a snapshot into a separate collection and retrieve a known image. **K1, K10, K11, K13**
- [ ] **SON-03 · Delivery demo.** Compare each changelog against delivered changes. Record release date, version, and migration steps for a release. Perform a documented clean setup. With two browsers exercise voice search, acquisition, repeat download, variation, new generation, and unique external contribution. Verify balance 100 → 99 → 96 → 93 → 94. Check private reference access and search for the published generation and contribution from the second browser. Record K1–K13 evidence with both repository revisions. **K1–K13**
