# BigData Analytics Dashboard

## 📝 Overview

This repository implements a **billion-row** event-log analytics dashboard with a uniform Go core for maximum throughput, plus a minimal Python slice to showcase data-science capabilities. Storage is handled by ClickHouse for true OLAP performance, while the front-end uses SvelteKit for a lightweight, reactive UI.

Summary of high-performance analytics:
- **Go** for backend data ingestion, transformation, and API services  
- **ClickHouse** as the analytics database (capable of handling billions of rows)  
- **Kafka** as a high-throughput, durable ingestion buffer  
- **Redis** as a low-latency cache for hot lookups and rate limiting  
- **SvelteKit** + **Chart.js** for a reactive, lightweight frontend  
- **Virtualized or paginated** event lists for efficient browsing  

---

## Why This Is Faster Than Most Solutions

1. **Compiled Go Services**  
   - **Native binaries** with zero interpreter overhead.  
   - **Static typing** and inlined calls reduce per-record CPU cost.  
   - **Goroutines & channels** provide efficient, zero-copy concurrency for parallel seeding and ingestion.

2. **ClickHouse for Analytics**  
   - Columnar storage and vectorized execution dramatically speed up aggregations and time-series queries.  
   - Designed for **massive scale**—can hold 1 billion+ rows with sub-second OLAP queries.  
   - MergeTree engine provides fast inserts and near-real-time reads.

3. **Kafka Ingestion Pipeline**  
   - Decouples data generation (seeder) from storage (ingest), smoothing spikes and providing durability.  
   - **Partitioned topics** let you scale consumers across multiple machines.  
   - **At-least-once delivery** ensures no data loss under failures.

4. **Redis Caching**  
   - Hot data (e.g. most recent minute’s event counts) is cached for microsecond reads.  
   - Eases load on ClickHouse for repeated dashboard refreshes.

5. **Frontend Efficiency**  
   - **SvelteKit** compiles to minimal vanilla JS; no virtual DOM.  
   - **Dynamic imports** of Chart.js and adapters, so initial bundle stays small.  
   - **Progressive skeleton loaders** keep users engaged while data streams in.

---

## Architecture & Data Flow

1. **Seeding**  
   - `seeder` (Go binary) generates synthetic `page_event` records (UUIDs, URLs, timestamps, metadata).  
   - Emits records into **Kafka** topic `page_events`.

2. **Ingestion**  
   - `ingest` (Go binary) consumes `page_events` from Kafka in batches.  
   - Writes batches into ClickHouse `analytics.page_events` table via the ClickHouse Go client.

3. **API Gateway**  
   - Built with **Chi router** in Go; listens on `:8080`.  
   - **Endpoints** (all `GET`):
     - `/metrics/overview`  
       Returns JSON `{ total_events, unique_users, events_last_hour }` by running three sub-queries in one ClickHouse request.
     - `/metrics/time-series?from=…&to=…&interval=…`  
       Streams back an array of `{ bucket: DateTime, count: UInt64 }` using `GROUP BY toStartOfInterval(...)`.
     - `/metrics/type-breakdown`  
       Returns a map `{ event_type: count }` via `GROUP BY event_type`.
     - `/metrics/events?cursor=…&limit=…`  
       Paginates raw events, filtering `ts > cursor`, ordered by `ts`, limit N.

   - **Streaming vs. Batch**  
     - Overview & breakdown endpoints return full JSON arrays (suitable for charting).  
     - Event pagination is **stateless paging** (cursor-based), not long-lived HTTP streams.

4. **Caching & Rate-Limit**  
   - **Redis** can be optionally introduced in the API layer to cache heavy queries (e.g. same time-series window) for sub-millisecond fetch.  
   - Also used for **rate-limiting** frontend polling (e.g. `X requests per second`).

5. **Frontend**  
   - **Environment**: reads `VITE_API_BASE` from `.env` to know where to call the API.  
   - **Data fetching** via SvelteKit `load` + client-side `fetch()`.  
   - **Skeleton loaders** during `loading` state, then dynamic import of Chart.js for charts.  
   - **Pagination** “Load more” button or optional virtualized list for event log.  

---


## 🔧 Technical Choices

- **Storage Layer: ClickHouse**\
  Columnar MergeTree engine, built for >10⁹ rows with vectorized execution, compression, TTL, and sharding.

- **Ingestion & Seeding: Go**

  - **High-throughput CSV seeder**: Go CLI using `gofakeit` to generate CSV shards and bulk-load via `clickhouse-client` (>200 k rows/sec).
  - **Real-time ingest**: Go gRPC server (grpc-go) publishing to Kafka.

- **Streaming Buffer: Kafka → ClickHouse**\
  Kafka for durable, replayable back-pressure; ClickHouse’s Kafka engine ingests directly.

- **Materialized Aggregates & Caching**

  - ClickHouse Materialized Views for continuous minute/hour buckets.
  - Redis Cluster caches summary keys (TTL \~30 s).

- **Backend Query Layer: Go**

  - `GET /metrics/overview` → Redis or continuous-aggregate tables (p95 < 5 ms).
  - `GET /events` → cursor-based JSON-Lines streaming (p95 < 50 ms for \~1 k rows).\
    Uses the official [clickhouse-go] client for native TCP performance.

- **Frontend: SvelteKit**

  - Zero-overhead RPC: fetch+Zod or lightweight tRPC proxy.
  - `svelte-virtual` for windowed lists, lazy-loaded chart components.

- **Python Slice**

  - **Seeder**: small `seed.py` with Faker+asyncpg to highlight Python skills.

---

## 🚫 Why Not PostgreSQL?

1. **Scale & Performance**: At ≈10⁹ rows, Postgres vacuum churn and partition maintenance hinder p95.
2. **OLAP Focus**: ClickHouse’s columnar storage and vectorized queries vastly outperform Postgres for analytics.
3. **Operational Simplicity**: ClickHouse automates compression and chunking—no manual partitioning.

---

## 🚫 Why Not Python (Core)?

1. **GIL & Pauses**: Python’s interpreter lock and GC can introduce unpredictable tail latency under heavy load.
2. **Throughput**: Go offers consistent multi-core performance and minimal startup overhead.

*Python remains in isolated slices only, ensuring the core data path is Go.*

---

## 📁 File Organization & Single-Responsibility Principles

- **One file = One role**: each file holds a single public function or component.
- **Keep files small**: target <200 lines; split when exceeding \~100 lines.
- **Minimal imports**: avoid coupling; if a file imports 3+ modules, consider splitting it.
- **Test per file**: one test file per package or component.

---

## 🏁 Getting Started (Linux All-in-One)

To streamline setup on a fresh Linux machine, we've provided a `setup.sh` script that installs prerequisites, configures services, seeds data, and starts all components.

### 1. Copy `env.example` 

```bash
cp env.example .env
cd frontend & cp env.example .env
```

Then update content according to your needs.

### 2. Make `setup.sh` executable

```bash
chmod +x setup.sh
```

### 3. Run the setup script

```bash
sudo ./setup.sh
```

This will perform the following steps:

1. **Install system packages**: ClickHouse, Kafka, Redis, Go, Node.js, Python3, pip
2. **Start services**: ClickHouse server, Kafka broker, Redis server
3. **Initialize database**: create `page_events` table in ClickHouse
4. **Build and launch Go services**: seeder, ingest, api-gateway
5. **Seed data**: launch Go seeder to insert 1 billion rows (configurable in `setup.sh`)
6. **Setup Python venv**
7. **Build and serve frontend**: install dependencies, build SvelteKit, and start dev server on port 5173

Once complete, access:

- **Dashboard:** [http://localhost:5173](http://localhost:5173)
- **Backend API:** [http://localhost:8080](http://localhost:8080)



