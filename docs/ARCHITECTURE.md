# qBitSlskd architecture plan

This document outlines a pragmatic refactor plan to modularize the current monolithic code into clear layers and packages, while keeping behavior intact. It’s designed for incremental adoption via small PRs.

## Goals

- Separate transport (HTTP, XML/JSON) from domain logic and external API clients
- Make slskd/qBittorrent compatibility logic testable via interfaces and mocks
- Centralize configuration, logging, and HTTP client timeouts
- Isolate caches/background jobs from handlers
- Keep optional AI functionality decoupled and easy to turn off

## High-level design

- Entry points (cmd): thin main that wires config, services, and HTTP server
- Adapters: HTTP handlers for QBT API and Torznab, mapping to domain models
- Services: domain logic orchestrating slskd client + caches + name resolver
- Clients: slskd HTTP client, with timeouts, context, and typed responses
- Models: shared domain types (Torrent, Category, SearchResult, etc.)
- Infra: config, logging, caches, middleware, HTTP server assembly

## Proposed folder layout

```
qBitSlskd/
├─ cmd/
│  └─ qbitslskd/
│     └─ main.go               # Bootstrap: load config, build services, start server
├─ internal/
│  ├─ config/
│  │  ├─ config.go             # Config struct, env parsing/validation
│  │  └─ validate.go
│  ├─ domain/
│  │  ├─ models.go             # Torrent, Category, Search*, AlbumName, etc.
│  │  └─ errors.go
│  ├─ clients/
│  │  └─ slskd/
│  │     ├─ client.go          # Interface + httpClient impl
│  │     └─ dto.go             # Wire types from slskd API
│  ├─ services/
│  │  ├─ torrent_service.go    # Build TorrentInfo from slskd download state
│  │  ├─ search_service.go     # Torznab search orchestration
│  │  └─ name_resolver.go      # Album name resolver (AI/heuristics)
│  ├─ adapters/
│  │  ├─ http/
│  │  │  ├─ server.go          # Router + middleware registration
│  │  │  ├─ middleware.go      # logging, recover, auth, CORS
│  │  │  ├─ qbittorrent/
│  │  │  │  ├─ handlers.go     # /api/v2/* compatible endpoints
│  │  │  │  └─ dto.go          # QBT WebAPI response DTOs
│  │  │  └─ torznab/
│  │  │     ├─ handlers.go     # /api?t=caps|search|custom_download
│  │  │     └─ dto.go          # XML structs
│  ├─ infra/
│  │  ├─ httpclient.go         # Shared http.Client with sane timeouts
│  │  ├─ logging.go            # Logger abstraction (stdlog/zap/slog)
│  │  ├─ cache.go              # album/search cache (TTL, size bounds)
│  │  └─ workerpool.go         # bounded concurrency for AI lookups
│  └─ translate/
│     ├─ qbittorrent.go        # Map domain -> QBT responses
│     └─ torznab.go            # Map domain -> Torznab feed
├─ pkg/                        # Optional: public utilities shared externally
│  └─ util/
│     └─ slice.go
├─ docs/
│  └─ ARCHITECTURE.md          # This file
└─
```

Notes:
- internal/ keeps implementation details private to this module
- cmd/qbitslskd contains only assembly/bootstrap; no business logic
- Existing files will move into the relevant internal/* packages with package renames

## Key interfaces (contracts)

These enable testing and future replacement of implementations.

- SlskdClient
  - GetDownloads(ctx, apiKey) ([]DownloadUser, error)
  - DeleteFile(ctx, apiKey, username, id string, remove bool) error
  - StartSearch(ctx, apiKey string, req SearchRequest) (searchID string, err error)
  - GetSearchState(ctx, apiKey, searchID string) (SearchState, error)
  - GetSearchResponses(ctx, apiKey, searchID string) ([]SearchResult, error)
  - DeleteSearch(ctx, apiKey, searchID string) error

- NameResolver
  - Resolve(ctx context.Context, path string) (album string, err error)

- Cache[K,V]
  - Get(K) (V, bool)
  - Set(K, V, ttl)
  - Delete(K)
  - PurgeExpired()

- TorrentService
  - ListTorrents(ctx, apiKey string) ([]Torrent, error)
  - DeleteByHashes(ctx, apiKey string, hashes []string, deleteFiles bool) error

- SearchService
  - Search(ctx, apiKey, query string, limit int) ([]DirResult, error)
  - GetDownloadDescriptor(ctx, cacheID string) (Descriptor, bool)

## Data boundaries

- domain/models.go defines canonical types: Torrent, DirResult, Descriptor, Category
- adapters/http/*/dto.go define wire types for QBT and Torznab only
- translate/* maps domain types to wire types; pure functions, unit-testable

## HTTP server and middleware

- Use chi or net/http; prefer chi for concise routes and middlewares:
  - Recoverer, RealIP, RequestID, Timeout, Logger
- Auth: SID cookie -> apiKey extraction in middleware; attach to context
- Content-type set by handlers; error helpers centralize status and payload

## Configuration and timeouts

- Config struct with env tags (envconfig) or viper. Validate on startup.
- Single shared http.Client with timeouts and keep-alives for slskd
- Feature flags: DELETE_SEARCHES, DOWNLOAD_AUDIO_ONLY, AI_ENABLED

## Caching and background jobs

- Album name cache: TTL + size bound; use ristretto or sync.Map + expiring entries
- Search cache: TTL; periodic sweeper goroutine or on-access cleanup
- Bounded concurrency for NameResolver to avoid API overrun (e.g., 10–20 workers)

## Error handling

- Consistent error values in domain; translate to proper HTTP codes in adapters
- QBT API nuances retained: some endpoints return 200 with specific bodies; captured in translator layer

## Testing strategy

- Unit tests:
  - translate/qbittorrent: mapping from domain -> TorrentInfo
  - translate/torznab: mapping to XML DTOs (golden files)
  - services: with mocked SlskdClient and NameResolver
- HTTP handler tests with httptest + fake services
- Integration tests (optional): spin up a fake slskd server with canned responses

## Migration plan (incremental)

1) Create internal/domain models and translate/qbittorrent mapping
   - Extract current TorrentInfo and move mapping logic out of handlers
   - Add unit tests for progress/eta/state fields

2) Introduce internal/clients/slskd client
   - Move slskd/main.go into typed client; add context + timeouts
   - Replace direct calls in handlers with client via service

3) Add services/torrent_service and services/search_service
   - Move orchestration (hashing, category join, speed/eta) into services
   - Keep handlers delegating to services only

4) Add adapters/http server and split handlers
   - Move qBittorrent and Torznab handlers into their own packages
   - Add minimal middleware (request logging, panic recovery)

5) Wire cmd/qbitslskd main.go
   - Build config, http client, caches, services, router; start server
   - Keep old main.go until parity is verified, then remove

6) Add NameResolver interface and AI resolver
   - Current ai.go becomes an implementation behind the interface
   - Add a no-op heuristic resolver when GEMINI_API_KEY is empty (feature flag)

7) Add caching and sweeper
   - Replace global maps with cache implementation
   - Add background goroutine with context for TTL eviction

8) Polish: config validation, structured logs, metrics hooks (optional)

## Quick wins you can do first

- Move constants and small helpers: util.go -> pkg/util or internal/translate helpers
- Add http.Client timeouts everywhere
- Centralize hash function and path splitting logic
- Add small tests around sha1 and isAudioFile edge cases

## Future enhancements

- Health and readiness endpoints
- Prometheus metrics for handler latency and slskd call counts
- Rate limiting to protect slskd and AI API
- Configurable concurrency and limits via environment

---

This plan keeps the behavioral quirks for Lidarr/QBT compatibility while making the code maintainable and testable. Adopt it in small, verifiable steps; each milestone should keep the server runnable and covered by basic tests.