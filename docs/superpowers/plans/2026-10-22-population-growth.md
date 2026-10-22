# Population growth implementation plan

**Goal:** Remove name service latency from births and immigration, and show demographic changes in Grafana.

**Approved scope:** The user accepted background name prefetch, a fallback name when the service is unavailable, and birth/death/immigration graphs.

**Architecture:** A single background worker fills a bounded channel with batches of names. Births and immigration take a name without waiting; the resident constructor supplies a unique ID-based fallback for an empty name. Successful demographic changes increment counters under the colony mutex, exposed by the existing collector.

**Constraints:** Work in the existing checkout to preserve the user's uncommitted implementation. No code comments or automated tests. Preserve all residents, IDs, saved progress, balance settings, database schema, and existing dashboard panels. No commits or destructive process termination.

- [x] Add a cancellable name pool: batches of 100, capacity 200, refill every 5 seconds below 100, one HTTP request at a time with a 5-second timeout. Remove network calls and the shared 10-second birth timeout from demographic loops.
- [x] Add ID-based fallback names and count only successful births, deaths, and immigration. Counters start at zero on process restart and are not historical database totals.
- [x] Expose `village_births_total`, `village_deaths_total`, and `village_immigrants_total` as Prometheus counters from one locked stats snapshot.
- [x] Add one Grafana comparison panel using rates per real minute, preserve existing panels, and document reset behavior.
- [x] Verify formatting, build, vet, JSON/config validity, and review changes independently.
- [ ] After a graceful simulator restart, observe the new live metrics and Prometheus queries.

**Review focus:** Empty pool and service failure must not suppress births. Queue memory and HTTP concurrency must stay bounded. Cancellation must stop refill and resident creation. Failed births and restored history must not increment counters. All demographic counts must share the colony lock; existing persistence must stay compatible.

**Verification:** `go build ./...`, `go vet ./...`, `git diff --check`, and `docker compose config --quiet` passed. Dashboard JSON parsed successfully and Grafana's dashboard API already returns the new demographic panel. The name service returned a valid batch of 100 names in a manual request; a preceding empty response is handled by the retry/fallback path. An independent static review found no important defects.

**Runtime status:** The live simulator still uses the previous binary and exposes only the previous village metrics. Computer Use rejected access to Terminal, so no process was stopped or restarted. The user must exit via Esc (normal final save) and repeat their existing launch command before live counter verification is possible.
