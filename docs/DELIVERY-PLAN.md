# MVP delivery plan and acceptance criteria

From a clean local installation, a user must be able to describe a need in Turkish by text or voice, find and acquire a suitable catalog image, or generate or modify one. An eligible shared image must become searchable. Balance, visibility, history, and file delivery must stay consistent throughout.

Unchecked boxes denote outstanding acceptance work. The [product plan](IMAVARA-MVP.md) defines behavior and model settings. The [service tasks](SERVICE-TASKS.md) assign owners, prerequisites, and implementation steps.

## Implementation sequence

| Stage | Work | Demonstrable result |
| --- | --- | --- |
| 1 | Prepare both repos, environment template, Compose network, volumes, migrations, and startup commands. Exercise seven model routes with real inputs. | Record per-model results, especially reference generation, embedding lengths, audio format, and Glimmer JSON. |
| 2 | Implement sessions, asset IDs, file rights, balances, ledger, and idempotency. | Verify private files, acquisition, and downloads with two browsers. |
| 3 | Initialize Qdrant and payload indexes before ingestion. Connect contributions, deduplication, Luna, queues, and workers. Ingest licensed seed images through the same path. | Trace one real upload to publication. Exercise rejection, duplicates, and redelivery. |
| 4 | Implement text/image search, strict filters, acquisition, rankings, and Discover. | Find a suitable image and download the acquired original in Discover. |
| 5 | Connect explicit generation and variation requests to Muse Image, Garage, and the ledger. Implement library, sharing, and withdrawal. | Deliver in the same request and trace the reference and charge. |
| 6 | Connect private MCP, LangChain, WebSocket, React chat, voice, history, and reference cards. | Spoken and written requests use the same product operations. |
| 7 | Complete dashboards, fault and retry checks, search evaluation, and a clean demo. | Record reproducible evidence for each criterion below. |

## Required acceptance checks

For each check record the command or user steps, expected and actual result, both repo revisions, and relevant request or asset IDs. Provider faults may be injected. Use real OpenRouter calls for live model acceptance.

### Local environment and services

- [ ] **K1 · Clean installation.** Start React and Compose from empty volumes using documented prerequisites. Verify the Go API, internal MCP, TypeScript agent, both workers, PostgreSQL, Qdrant, RabbitMQ, Garage, Prometheus, and Grafana. Restart and verify that data and migrations survive. The browser uses web/API, and private storage, MCP, and data services stay internal. Keep the provider key server-side. Run locally without a Sentry key.
- [ ] **K2 · Model contracts.** Record a successful request for each exact model ID in the plan. Verify Ember tools and streaming, Turkish transcription, fresh and reference-based Muse images, Luna decisions, valid versioned Glimmer JSON, and finite Qwen/Gemini vectors of the configured lengths. Exercise server profiles. Leave a check open if a required provider capability is unsupported.

### User experience

- [ ] **K3 · Chat and tools.** The opening screen shows the wordmark, one input, microphone, and Meta Muse credit. First submission opens chat. An ambiguous need elicits one short question. A clear search returns real API cards over internal MCP, up to five by default or ten with explicit `top_k`. WebSocket carries text and actual tool events. Empty search incurs no generation or charge. “Find something like the second one” uses the second card for a free reference search.
- [ ] **K4 · Voice.** Record on the landing page and in chat. A single Send stops recording and shows transcription progress. Preserve the aspect ratio, direction, and selected-card meaning in a Turkish recording. Combine transcript, draft, and reference in one message and one agent turn. Double taps and reconnects produce the same outcome. Cancellation, denied permission, empty transcripts, and errors preserve drafts. Raw audio stays out of persistent storage.
- [ ] **K5 · Discover and history.** Run filtered text and reference searches without an agent turn. Cards show the preview, ratio, and dimensions. After refresh, create a conversation, reopen an old one in card order, delete it, and return to Library. Deleting a conversation preserves acquired assets.
- [ ] **K6 · Generation and variations.** For an explicit request, store one original in Garage and deliver in the same request. For “remove the cup in the second image,” inspect the reference sent to the provider, the requested edit, and preserved framing. Display actual dimensions. Automatically submit for admission and find the published result from a second browser.

### Credits, access, and catalog

- [ ] **K7 · Balance.** Grant 100 credits once per browser. Chat, voice, search, and preview cost zero. First acquisition costs 1, repeat download zero, delivered generation or variation 3. Reward the first publication of a unique external contribution with 1. Insufficient balance must prevent a paid generation call. Test double clicks, repeat HTTP keys, two concurrent requests for the last three credits, and definite failure. Reconcile ledger, reservations, and available balance. An ambiguous provider response must not cause a second generation.
- [ ] **K8 · Access and withdrawal.** Use two independent browsers to verify private chats, assets, original files, and balances, including direct ID access. Ordinary withdrawal closes new search and acquisition while prior holders retain downloads at the persistent URL. Catalog rejection preserves owner delivery.
- [ ] **K9 · Batch contributions and admission.** Mix original files, byte-identical duplicates, similar but distinct images, and a rule violation. Show each outcome. Verify full-hash duplicate detection, sharing permission, usage terms, and Luna acceptance. Distinguish rejection, review, and provider error. Concurrent copies yield one catalog entry and at most one reward.
- [ ] **K10 · Analysis and queues.** Carry asset ID, object key, and version in RabbitMQ. Store verified versioned Glimmer JSON privately. Index its text and filters plus a vector from the image. Stop a worker after writing but before ACK and restart it. Verify completion, file durability, one point and one reward. Persistently invalid results reach the failure queue after bounded retries. Withdrawal during processing must prevent publication.

### Search and operations

- [ ] **K11 · Collection and filters.** Confirm one point with cosine `text[1024]` and `image[768]`, 16 shards, one node, replication factor 1, binary quantization, and original-vector rescoring. Create all planned payload indexes before ingestion. Verify every hit against visibility, license, aspect, and other exact filters. Strong results obey the field-specific threshold. With nearest fallback enabled, return at most one eligible lower-threshold candidate when strong results are empty. Restore a snapshot to another collection and run a known query.
- [ ] **K12 · Search quality.** Label about 100 Turkish/English text queries and additional image references. Split calibration and acceptance cases. Select separate thresholds for text and image search. Record thresholds, sample sizes, and model and collection versions with the held-out results.
- [ ] **K13 · Statistics and diagnosis.** Return the same image in two searches and acquire it once. Verify two returns and one first acquisition for the time window. Repeat downloads and viewing the popular list must preserve the counts. Check recent and popular sort orders in Discover. Open Prometheus metrics and Grafana dashboards for all five app services. Follow one request or job through shared logs. Observe search latency, empty results, provider cost/errors, queue age, and time to publication.

## Model quality targets

On held-out answerable requests, aim for a labeled relevant image in the first five at least 80% of the time. On unanswerable requests, aim for at least 90% correct empty results. Report the two rates separately for text and image references. Include positive and negative cases in each group and report language distribution, filter combinations, counts, and failures. Calibrate cosine thresholds separately from these product targets.

Compare binary retrieval to exact full-precision search. Record p95 latency, OCR/composition errors, and total cost per valid Glimmer analysis for the first 100 representative images, including retries. Use measured local hardware and provider behavior to inform latency and budget decisions.

## Final demo and release decision

Start with 100 credits in a fresh browser. Ask by voice for a 4:3 café image with room for a title on the left. Seed at least two suitable images beforehand. Acquire the second suitable card and download it twice (99). Request a variation of that card (96). Generate a new image (93). From a second browser, find the accepted generation and check original-file entitlement. Submit a separate, licensed, unique external photo from the first browser. After publication its balance reaches 94. Find the contribution from the second browser. Reopen the original conversation and compare card order, library, balance, and ledger.

Close the MVP when K1–K13 and the final demo pass on the same release. Record outstanding behavior, reproduction steps, and impact for each open check. Keep estimates labeled until measured.

### Local verification record · November 24, 2026

- A live API and PostgreSQL check with two cookies confirmed 100 credits per browser, one initial grant, session continuity, and HTTP 404 when requesting another browser's private chat.
- On a separate PostgreSQL test database, `TestSessionAndConcurrentReservations` passed. Two concurrent attempts to reserve the last three credits produced one reservation. Reusing the key created none. A definite failure released the funds.
- The previous distributed profile reported two peers, 32 local and 32 remote shard copies, replication factor 2, two cosine vectors, and 17 payload indexes. Garage initialization exited successfully. This run predates the current K11 profile.
- These observations are partial evidence for K1, K7, K8, and K11. At the time of this record, live model, end-to-end, and retrieval-quality acceptance remained open.
