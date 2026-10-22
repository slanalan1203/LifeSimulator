# Monitoring Stack Implementation Plan

> **For agentic workers:** Use superpowers:executing-plans or superpowers:subagent-driven-development to execute the scoped tasks below.

**Goal:** Show the running Go simulator's existing metrics in a persistent, locally accessible Prometheus/Grafana dashboard.

**Architecture:** Keep the simulator in its current terminal process. Prometheus runs in Docker Desktop and scrapes `host.docker.internal:8080/metrics` every five seconds. Grafana uses a provisioned Prometheus data source and dashboard.

**Tech Stack:** Existing Go Prometheus collector, Docker Compose, Prometheus, Grafana.

**Spec:** Agreed chat scope: one useful dashboard for population, food, wood, houses, Go memory and goroutines.

## Global Constraints

- Preserve the running village, PostgreSQL service and existing database volume.
- Add no tests or explanatory code comments, following the user's preferences.
- Bind monitoring ports to localhost and persist monitoring data in separate named volumes.
- Use only existing metrics; do not change simulation rules or restart the user's simulator.

## Review Focus

- Docker Desktop must reach the host process, which listens on localhost.
- Missing simulator data must remain visibly missing rather than be converted to zero.
- Grafana queries must use the provisioned data source UID and correct metric labels.
- Dashboard and data source provisioning must work on an empty Grafana volume.
- Startup commands must leave the existing PostgreSQL container and data intact.

## Task 1: Scraping and container configuration

- [x] Verify published image tags, available ports and host connectivity.
- [x] Extend `docker-compose.yaml` with Prometheus and Grafana, persistent volumes, localhost ports and health checks.
- [x] Create `monitoring/prometheus/prometheus.yml` with job `life-simulator`, five-second scraping and a four-second timeout.
- [x] Validate Compose and Prometheus configuration; start only the monitoring services.

## Task 2: Provisioned dashboard

- [x] Create Grafana provisioning files under `monitoring/grafana/` and a classic JSON dashboard with UID `life-simulator`.
- [x] Use data source UID `prometheus`, URL `http://prometheus:9090`, and job selector `job="life-simulator"`.
- [x] Include current values and historical charts for the agreed metrics, plus scrape availability.
- [x] Verify the real data source, dashboard and Prometheus target through their HTTP APIs and browser.

## Task 3: Usage documentation

- [x] Document startup, URLs, local credentials, retention and simulator-down behavior in `README.md`.
- [x] Run relevant static and configuration checks, inspect the dashboard, and report the verified result.

## Verification Record

- Compose configuration and `promtool check config` passed.
- Both monitoring containers are healthy; Prometheus reports target `life-simulator` as `up`.
- Grafana's data source and nine-panel dashboard were provisioned successfully on a new volume.
- Live values and historical charts were inspected in the browser at narrow and wide widths.
- After restarting only Prometheus and Grafana, a pre-restart historical population sample remained queryable and the dashboard remained provisioned.
- Focused independent review found no important issues. Existing Go collectors were not changed by this phase.
