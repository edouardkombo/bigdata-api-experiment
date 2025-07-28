# BigData Analytics Dashboard

## 📖 Overview

This repository implements a **billion-row** event-log analytics dashboard with a uniform Go core for maximum throughput, plus a minimal Python slice to showcase data-science capabilities. Storage is handled by ClickHouse for true OLAP performance, while the front-end uses SvelteKit for a lightweight, reactive UI.

Summary of high-performance analytics:

- **Go** for backend data ingestion, transformation, and API services
- **ClickHouse** as the analytics database (capable of handling billions of rows)
- **RabbitMQ** as the ingestion buffer (replacing Kafka)
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
   - Designed for **massive scale** — can hold 1 billion+ rows with sub-second OLAP queries.
   - MergeTree engine provides fast inserts and near-real-time reads.

3. **RabbitMQ Ingestion Pipeline**

   - Decouples data generation (seeder) from storage (ingest), smoothing spikes and providing durability.
   - Queued ingestion allows replay, burst handling, and safer failure recovery.
   - Simplified routing using one durable queue (`events`).

4. **Frontend Efficiency**

   - **SvelteKit** compiles to minimal vanilla JS; no virtual DOM.
   - **Dynamic imports** of Chart.js and adapters, so initial bundle stays small.
   - **Progressive skeleton loaders** keep users engaged while data streams in.

---

## Architecture & Data Flow

1. **Seeding**

   - `seeder` (Go or Python) generates synthetic `page_event` records.
   - Emits records to **RabbitMQ** queue `events`.

2. **Ingestion**

   - `consumer` (Go) listens to RabbitMQ and writes batches into ClickHouse `analytics.page_events`.

3. **API Gateway**

   - Built with **Chi router** in Go; listens on `:8080`.
   - **Endpoints**:
     - `/metrics/overview` → Total, unique users, first/last timestamps
     - `/metrics/time-series?from=...&to=...&interval=...` → Aggregated event counts per time bucket
     - `/metrics/type-breakdown` → Counts per `event_type`
     - `/metrics/events?user_id=...&event_type=...` → Filtered recent event rows

4. **Frontend**

   - Reads `VITE_API_BASE` from `.env`
   - Uses SvelteKit + `Chart.js` + dynamic imports
   - Built-in loaders and modular chart components

---

## 🛠️ Technical Choices

- **Storage Layer: ClickHouse**\
  MergeTree engine with vectorized execution and built-in TTL, partitions, and compression.

- **Ingestion & Seeding: Go**\
  High-throughput seeding tools and durable RabbitMQ pipelines.

- **Queue Layer: RabbitMQ**\
  Simpler, pluggable, and easier to manage than Kafka for this scale.

- **Frontend: SvelteKit**\
  Ultra-light reactivity and dynamic chart rendering.

- **Python Slice**\
  Included only for seeding speed comparison and async data simulations.

---

## 🧾 File Organization

- `backend_go/cmd/producer` → HTTP API → RabbitMQ
- `backend_go/cmd/grpcserver` → gRPC API → RabbitMQ
- `backend_go/cmd/consumer` → RabbitMQ → ClickHouse
- `backend_go/clickhouse/init.sql` → Schema
- `backend_go/proto/event.proto` → Protobuf definitions
- `backend_python` → Python Seeder
- `frontend` → Svelte frontend

---

## 🧪 Getting Started (Linux All-in-One)

### 1. Copy `.env` files

```bash
cp env.example .env
cd frontend && cp env.example .env
```

### 2. Make `setup.sh` executable

```bash
chmod +x install_packages.sh
```

### 3. Run the setup

```bash
sudo ./setup.sh 100000
```

This will:

- Install RabbitMQ, ClickHouse, Go, Python, and frontend deps
- Initialize the DB schema
- Launch all services
- Seed data using Go or Python
- Serve dashboard on [http://localhost:5173](http://localhost:5173)

---

## 🔪 Test It

### HTTP

```http
POST http://localhost:8080/events
Content-Type: application/json

{
  "user_id": "abc-123",
  "event_type": "click",
  "url": "https://example.com/page",
  "referrer": "https://google.com"
}
```

### gRPC

Call `PublishEvent` on port `50051` using `grpcurl`, Postman, or a generated client.

---

## 📊 Architecture Flow

```
HTTP / gRPC Producer
        │
        ▼
    RabbitMQ (queue: events)
        │
        ▼
  Consumer (Go)
        │
        ▼
 ClickHouse (table: analytics.page_events)
```

## 🛋️ Optional Cleanup

```bash
sudo lsof -ti:8080 | xargs sudo kill -9
sudo lsof -ti:50051 | xargs sudo kill -9
```



