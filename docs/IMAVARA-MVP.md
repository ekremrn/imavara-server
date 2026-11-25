# Imavara local MVP product and architecture plan

This document defines the product scope and target architecture. The service boundaries, routes, directories, and validation work below describe how the design fits together. The delivery plan tracks evidence of completion.

## 1. Product and success criteria

Start by searching. Generate an image when the catalog cannot meet the brief. A user describes a need, reviews existing AI images, and can acquire an original, vary a close match, or generate a new image. Once accepted and published, an image can serve other users.

Measure how often a real brief ends with an existing image, the fraction of searches with no results, transitions to generation, and cost per acquired image. Seed the catalog with images that have known usage rights, varied compositions, and different aspect ratios.

For example, a café owner asks for a warmly lit website cover. The agent asks where the headline will go. The owner answers “on the left,” sees real cards, and can acquire the second result for 1 credit or ask for a 3-credit variation that removes the cup while preserving the framing. Reopening the chat the next day preserves card order.

## 2. Local deployment boundary

React runs locally. Docker Compose runs the API, internal MCP server, web agent, two workers, and data services. Model traffic goes to OpenRouter over the internet. Credentials stay on the server.

A persistent random HttpOnly cookie identifies an anonymous browser in PostgreSQL. HTTP and WebSocket use that same context for chat, balance, library, and downloads. The agent cannot choose a browser identity. Imavara's LangChain agent is the first MCP client, on the private Docker network. Access for another private-network client is a future deployment choice.

Keycloak, OAuth, roles, SMTP/Mailpit, Kubernetes, Netlify deployment, production operations, and credit sales are outside the local MVP. Revisit identity, deployments, and Sentry for production. SVG components, video, and other media types are later phases.

## 3. Interface and conversation

**Landing.** A small “imavara.” wordmark and one input field let the user type or attach a reference. On the first message, the field moves down and the conversation opens. Preserve keyboard focus and respect reduced motion.

**Voice.** Offer recording on the landing screen and in chat. Request permission on first use. Show elapsed time, Send, and Cancel. Send stops recording, displays the transcription state, combines the transcript with any typed draft, and submits one message with the selected reference. Show search progress when the actual tool starts. Preserve the draft on cancellation, denied permission, empty transcription, or failure.

**Chat.** Ask one short question when purpose, ratio, style, framing, or copy space materially affects the result. Search when the brief is clear. Keep real cards in order with the conversation. Users can start a new chat or reopen an old one. Deleting a chat and removing a library image are separate operations. [Agent system prompt](../agent/prompts/system.tr.txt)

**Discover.** Direct text and reference searches, filters, recently published images, and most-acquired images are available without an agent turn. Cards retain the original framing and show preview, dimensions, aspect ratio, and actions to acquire, find similar images, or generate a variation. Batch contributions report an outcome per file.

**Visual design.** Use warm background `#F7F6F2`, white surfaces, primary text `#202522`, secondary text `#58635D`, teal accent `#176454`, and selected surface `#E3EFE9`. The lowercase wordmark ends with a period. The design proposes DM Sans and optionally Instrument Serif for the opening question. A secondary footer credits Meta Muse for chat, image generation, and analysis.

## 4. Browser credits and access

| Operation | Credits | Accounting |
| --- | ---: | --- |
| First browser session | +100 | One opening grant |
| Chat, voice, brief refinement, search, preview | 0 | Empty searches are free |
| First acquisition | −1 | Once per browser and asset |
| Repeat download | 0 | Same entitlement and persistent URL |
| New image or variation | −3 | Charge once on delivery. No separate source acquisition fee |
| First publication of a unique external contribution | +1 | Once per distinct contribution |

Credits are product units. Provider spending is measured separately. Tabs and conversations in one browser share a balance. Clearing the cookie, using private browsing, or switching profiles creates a new demo session. Previous conversations remain associated with the old session.

The API records balance changes transactionally. Reserve 3 credits before generation, settle once the stored original can be delivered, and release on definite failure. Operation keys protect retries of HTTP requests and WebSocket messages. Concurrent requests cannot both spend the last available credits.

## 5. Services and models

| Process | Responsibility |
| --- | --- |
| Go API | Browser context, HTTP/WebSocket, transcription, search, acquisition, generation, contributions, statistics, and credits |
| Go MCP | Forward internal tool calls to their API operations with the same browser context |
| TypeScript web agent | Run the LangChain conversation and MCP tools. Stream real cards, tool events, and response text |
| Go image-extract-worker | Consume analysis jobs and extract validated descriptions and attributes |
| Go index-worker | Normalize analysis, write vectors and filters, and complete publication |

All seven model IDs and their roles appear in the server profiles in section 12.

## 6. End-to-end operations

**Search.** Embed text with Qwen and external reference images with Gemini. Reuse a stored vector for an accessible catalog reference. Apply visibility, license, exact filters, and the appropriate similarity threshold. API and MCP accept integer `top_k` from 1 to 10, default 5. Strong matches appear in `items` with previews, file properties, and usage information. When `include_nearest: true` and no strong match exists, a single eligible candidate above the lower cutoff can appear in `nearest`. MCP and web request this behavior. Responses expose `score_threshold`, `candidate_threshold`, `metric`, and `vector`. Cards expose `match.score`, `match.percent`, and `match.level`. The percentage is a rounded cosine score. An empty search succeeds without triggering paid generation.

**Acquisition.** Record the first acquisition and credit movement together. Return a persistent Imavara URL. Check rights when serving the original from Garage/S3. Ordinary withdrawal closes new searches and acquisitions while preserving previous download entitlements. Rights-related removal is handled separately.

**Generation.** Submit the prompt, ratio, and accessible reference to Muse Image. Store the original in S3 before delivery in the same request. Automatically queue every generation and variation for catalog admission. Publication makes an accepted image searchable across browsers. A catalog rejection preserves delivery to the owner. Allow up to two safely retryable transient failures within the request. An ambiguous provider outcome must not start a second generation.

**External contributions.** The browser may catch repeated selections early. The API validates the actual file, reads dimensions, and hashes the bytes with SHA-256. Look up the content-derived Qdrant point ID and compare the full hash. Visual embeddings can identify similar candidates. Small meaningful variations remain distinct. Concurrent uploads and retries must yield at most one catalog record and reward.

**Admission.** Evaluate external contributions and generated candidates with Luna. Version the questions for intrusive watermarks, legible private information, severe damage, and unsuitable content. Map rule scores to acceptance, rejection, or review with reasons. Track provider failure separately and record sharing permission and usage terms.

**Extraction and indexing.** Persist the admission and pending job together. RabbitMQ carries `asset_id`, `object_key`, and version. The extraction worker reads the original from S3, validates Glimmer JSON, and stores a versioned result. Acknowledge the first delivery after the broker confirms the index message. The index worker prepares embeddings and filters, checks current sharing state, and writes Qdrant. Acknowledge after durable publication and any external contribution reward. Replaying the same asset version must be safe. Inspect repeated failures in the failure queue.

**Voice.** Send the recording to the Go API in the browser session. Validate format and request size, obtain a transcript from OpenRouter, and return the text. Raw audio exists only for transcription. Auto-submit a nonempty transcript over the existing chat WebSocket, creating a conversation on the landing screen. Preserve the submission ID across double taps and reconnects. Written and spoken messages have identical tool and credit rules.

**Live chat.** React uses WebSocket to reach the API. The API streams events from the private agent. The browser displays actual tool starts, cards, text chunks, delivery, and errors. Message IDs, paid-operation keys, and event order distinguish retries. On reconnect, reload completed messages, cards, and operation state from PostgreSQL.

## 7. Data, filters, and measurement

- **Garage/S3** separates originals, public previews, and private analyses. Versioned JSON supports per-asset inspection and reindexing. Parquet is an option for batch analytics. A reference and hash suffice for normal processing. A standalone dataset export may also include image bytes. Store Qdrant snapshots separately.
- **Qdrant** holds one point per published image with two named vectors, `text` and `image`. Choose model, dimensions, and score threshold per field. Use binary quantization with original-vector rescoring. Create payload indexes before the initial batch.
- **Glimmer** supplies descriptions, subjects, objects, people presence, scene, perspective, light, mood, style, colors, subject placement, copy space, visible text, tags, and potential uses. Controlled fields support uncertainty.
- **File metadata** supplies width, height, ratio, format, alpha, byte size, and hash. Application records supply visibility, origin, parent asset, model, sharing, and license. Search filters include aspect, resolution, style, color, copy space, people, text, visibility, and license.
- **PostgreSQL** stores sessions, ledger entries, acquisitions, conversations, messages, card order, publication, and search events. Record one search ID and one ranked return per asset in that search. First acquisition is unique per browser/asset pair. Aggregate returns and acquisitions over a date range for separate popularity lists. Repeat downloads and viewing a popularity list do not increment either count.
- **Observability** uses shared Go logging, request and operation IDs, Prometheus, and a Grafana dashboard per service. Track latency, empty searches, provider errors and costs, queue age, and time to publication. Local operation supports an empty Sentry DSN.

## 8. Interface contracts

The following routes describe the MVP contract. `/me` refers to the current browser. API and MCP share prices and operation semantics.

| API | Purpose | Consumer or MCP tool |
| --- | --- | --- |
| `GET /v1/me`, `GET /v1/pricing`, `GET /v1/me/credits/transactions` | Balance, prices, history | Web, `get_balance` |
| `POST /v1/search` | Free text or reference search | `search_images` |
| `GET /v1/assets`, `GET /v1/assets/{id}` | Gallery and asset details | Web, `get_image` |
| `POST /v1/assets/{id}/acquire`, `GET /v1/assets/{id}/original` | Acquire and download original | `acquire_image`, web |
| `POST /v1/generations` | Generate or vary a reference | `generate_image` |
| `POST /v1/contributions` | Per-file submission and outcome | `contribute_image`, web |
| `GET /v1/me/assets`, `PATCH /v1/assets/{id}`, `DELETE /v1/assets/{id}` | Library, sharing, removal | Web |
| `GET /v1/stats/assets` | Period-filtered `acquisitions` or `search_returns` | Web |
| `POST /v1/audio/transcriptions` | Multipart recording in, `{"text":"…"}` out. Auto-submit the text to chat | Web |
| Conversation create/list/open/delete and WebSocket | Saved history, new messages, live events | Web agent entry point |

## 9. Repositories and code ownership

The web repository owns the shell in `src/app`, product screens in `src/features`, reusable cards in `src/components`, and HTTP/WebSocket integration. The server's single Go module builds independent binaries under `cmd/{api,mcp,image-extract-worker,index-worker}`. Domain packages under `internal` implement catalog, search, credits, generation, chat, extraction, and indexing. Shared integrations live under `internal/platform`.

The TypeScript LangChain service, MCP client, system prompt, package manifest, and Docker image live in `agent/`. See `contracts/`, `migrations/`, `deploy/`, and `docker-compose.yml` for schemas, database migrations, Qdrant, and infrastructure settings. The API owns credit and file-access decisions. Versioned contracts let front and back ends evolve independently.

## 10. Network and storage contract

The browser reaches the Go API through Vite's same-origin proxy for `/v1` and WebSocket. Compose maps the API's container port 8080 to localhost port 18080 by default. In a trusted LAN demo, expose the web entry point and the API address only if needed. API and workers access Garage over the internal network, e.g. `http://garage:3900`.

| Traffic | Path |
| --- | --- |
| Preview and original | Use the card's preview URL. `GET /v1/assets/{id}/original` checks entitlement before streaming the file. |
| Batch contribution | Submit each file to the API, validate format, hash it, and store it in S3. |
| Tools and queues | MCP, agent, RabbitMQ, Qdrant, and PostgreSQL communicate internally. Administrative ports are separate from the product entry point. |
| Provider image input | Read from private storage and send a supported inline representation over outbound HTTPS. RabbitMQ retains object keys. Verify inline support during integration. The provider cannot resolve local Docker URLs. |
| Optional media host | A read-only media address can serve public previews. If a provider needs a URL for private input, use a temporary object-specific read URL. Keep S3 administration, writes, private originals, Parquet, and backups internal. |

## 11. Qdrant collection contract

Bootstrap the collection, both named vectors, and payload indexes before ingesting images. The filtered HNSW graph benefits from indexes present at creation. Verify existing schemas at startup. For incompatible schema revisions, plan to populate and validate a new collection before moving the read alias. [Qdrant indexing](https://qdrant.tech/documentation/manage-data/indexing/)

| Field | Type | Purpose |
| --- | --- | --- |
| `owner_browser_id` | UUID | Ownership filtering if private-library indexing is added |
| `visibility`, `catalog_status`, `license_scope` | Keyword | Restrict public search to published, shared, licensed records |
| `aspect_ratio`, `orientation`, `mime_type` | Keyword | E.g. 4:3, landscape, image/png |
| `style`, `theme`, `color_family`, `negative_space` | Keyword, including arrays | Controlled visual and composition filters |
| `width`, `height` | Integer | Minimum or range in pixels |
| `has_people`, `has_text`, `has_alpha` | Boolean | Omit undetermined values |
| `published_at` | Datetime | Date ranges and recent publications |

Keep the full SHA-256 in the payload. Identify exact duplicates using a hash-derived UUID and full hash comparison. Keep private file existence confidential and acquisition records independent of catalog IDs.

**Search profile.** Both `text[1024]` and `image[768]` use float32 and cosine distance. Initial settings use original vectors on disk and one-bit binary representations in memory, HNSW `m=16`, `ef_construct=100`, query `hnsw_ef=128`, `oversampling=3.0`, and `rescore=true`. Select the appropriate field, apply payload filters, rescore originals, and enforce thresholds and `top_k`. [Qdrant quantization](https://qdrant.tech/documentation/manage-data/quantization/)

**Calibration.** Use about 100 labeled Turkish/English queries plus image references, including absent subjects, left/right copy space, and strict filter combinations. Compare to exact search, measure top-five hits, false returns on unanswerable queries, and p95 latency. Tune separate text and image thresholds on calibration cases and record held-out results with model and collection versions. Search relevance, duplicate detection, and catalog admission are separate decisions.

**Topology.** One local node, one volume, 16 shards, replication factor 1. Restore a snapshot into a separate collection and verify retrieval.

## 12. Server model profile v1

Clients supply the brief, filters, reference, and desired ratio. The server owns provider selection, sampling, reasoning, dimensions, and model IDs. Track profile version, actual model, latency, and cost.

| Model | Settings and checks |
| --- | --- |
| `meta/muse-ember` | Chat and tool selection. Temperature 0.3, low reasoning effort, automatic tool choice, streaming on. Keep effort fixed without a fixed thinking-token budget. Stream real tool events and user-facing text. [Endpoints](https://openrouter.ai/api/v1/models/meta/muse-ember/endpoints) · [Reasoning levels](https://docs.cloud.google.com/gemini-enterprise-agent-platform/models/partner-models/meta/muse-ember-1-0) |
| `openai/gpt-transcribe` | Transcribe in the original language through OpenRouter's `POST /api/v1/audio/transcriptions` using `input_audio.data` and `input_audio.format`. Measure recording duration, latency, and cost. Test a Turkish browser recording with ratio and direction words. [Model](https://openrouter.ai/openai/gpt-transcribe) · [API](https://openrouter.ai/blog/tutorials/transcription-on-openrouter/) |
| `meta/muse-image` | One image per request. Send one reference for a variation and specify what to preserve. Default to 1:1 for new images or the reference ratio for variations. Use provider defaults for temperature, top_p, seed, quality, and thinking. Verify reference and ratio support and report the actual file dimensions and any discrepancy. [Model](https://openrouter.ai/meta/muse-image) · [Image API](https://openrouter.ai/blog/announcements/image-api/) |
| `qwen/qwen3-embedding-8b` | Text retrieval at 1024 dimensions, one quarter the float storage of 4096. Use a fixed retrieval instruction for queries and description, composition, and attributes for documents. When receiving all 4096 dimensions, verify the MRL truncation to the first 1024 and L2 normalization. Share transformations between API and index worker. [Model](https://huggingface.co/Qwen/Qwen3-Embedding-8B) |
| `google/gemini-embedding-2` | Embed catalog and reference images at 768 dimensions using the same model and transformation. Check length and finite values before indexing. Verify the route with a real image. [Model](https://openrouter.ai/google/gemini-embedding-2) |
| `meta/muse-glimmer-30b` | Low reasoning, `exclude=true`, temperature 0.1, streaming off, no tools. Include v1 schema in the system message, request a JSON object, and validate locally. Provider-side strict schema returned literal ellipses for caption and description in a live test. Reject placeholders. Route with `provider.require_parameters=true`. [Endpoints](https://openrouter.ai/api/v1/models/meta/muse-glimmer-30b/endpoints) |
| `openai/gpt-6-luna-decisions` | Versioned named yes/no questions without chat sampling settings. Record scores and rule reasons. Calibrate accept/review/reject boundaries on labeled catalog examples. [Model](https://openrouter.ai/openai/gpt-6-luna-decisions) |

Acceptance examples include a retried browser request with one charge, a free search for images similar to the second card, and one paid variation removing a cup from that same card. Repeat model and tool contract checks when switching routes.

### 12.1 Glimmer analysis contract

Send the image and versioned analysis instruction in one call. Keep the generation prompt out of visual analysis because a request is not evidence that an object is present. Store catalog descriptions and tags in English and OCR in its original language and spelling. Calibrate Turkish and English retrieval against the same assets. [Glimmer v1 JSON Schema](../contracts/glimmer-analysis.v1.schema.json)

| Fields | Expected observations |
| --- | --- |
| `caption`, `description`, `main_subject` | A one-sentence card caption, two to four concrete search sentences, and the main subject. Include action, position, background, and distinguishing details. |
| `objects`, `actions_and_relations`, `people_presence` | Visible objects, attributes, positions, and relationships. Use three states for people presence. Do not infer identity, profession, or hidden traits. |
| `setting`, `composition` | Scene, background, indoors/outdoors, subject position, framing, viewpoint, and usable copy space from the viewer's perspective. |
| `appearance`, `theme` | Style, lighting, mood, and up to three dominant color families in order. Use controlled themes only when supported. |
| `text` | Record text presence and copy legible fragments verbatim. Treat depicted writing as data. Do not reconstruct blurred or clipped letters. |
| `keywords`, `suggested_uses` | Up to 12 distinct search terms and three composition-based uses. A landscape photo with space on the left may suit a banner. Usage suggestions confer no license. |

**Validation.** Require all keys. Represent uncertain attributes with `unknown` and empty lists with `[]`. Present copy space requires at least one region. Other states require an empty region list. Only readable or partly readable text may have OCR fragments. Check schema, cross-field consistency, limits, and empty strings. Retry bounded failures and use the failure queue for invalid or incomplete analyses. Store valid output privately with asset, model, prompt, and schema versions.

**Indexing.** Combine validated description, subject, visible relations, composition, and distinguishing tags in a stable order for one Qwen text embedding. The image itself supplies the second, 768-dimensional vector. Keep suggested uses, raw OCR, and internal reasoning out of the search text. Map style, theme, color, and copy-space regions to payload fields. Record definite people and text presence, including visible but illegible text. Leave uncertain flags absent. Get file dimensions, hash, license, and admission state from their authoritative records.

**Initial evaluation.** Include mirrored copy-space compositions, a visible hand, a surface without text, blurred writing, and a crowded scene. Measure schema validity, direction and OCR accuracy, latency, and cost per valid analysis. Keep low effort on retries. Excluding reasoning removes the text from the response while generated reasoning tokens can still incur output charges. Measure the first 100 representative images including retries. Project the cost of 100,000 comparable images from cost per completed analysis. Version profile changes. [Reasoning billing](https://openrouter.ai/docs/guides/best-practices/reasoning-tokens)

## 13. Architecture diagrams and open decisions

[Local API](architecture/api.png) · [Internal MCP](architecture/mcp.png) · [Extraction](architecture/extract.png) · [Indexing](architecture/index.png) · [Web agent](architecture/agent.png) · [System overview](architecture/overview.png)

Validate provider dimensions, binary quantization impact, and both similarity thresholds on the first catalog. The original 0.80 score suggestion is provisional. Tune Luna rules and review boundaries with examples and evaluate near duplicates separately. Private-library semantic indexing, chat retention, seed collection sourcing, end-to-end inline provider support, and the measured provider budget remain decisions to validate.

See the [delivery plan and acceptance criteria](DELIVERY-PLAN.md) for sequencing, completion conditions, and the final demo.
