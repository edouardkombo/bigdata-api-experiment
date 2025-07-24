#!/usr/bin/env python3
import os
import json
import sys
from uuid import uuid4
from faker import Faker
from datetime import datetime, timedelta
from clickhouse_driver import Client

def generate_events(total):
    fake = Faker()
    for _ in range(total):
        # generate a Python datetime, not an ISO string
        ts = fake.date_time_between(start_date='-7d')
        yield (
            str(uuid4()),
            str(uuid4()),
            fake.random_element(["page_view","click","scroll"]),
            fake.uri(),
            fake.uri(),
            ts,  # datetime object
            json.dumps({"x": fake.random_number(), "y": fake.random_number()})
        )

def seed(total, batch_size=10000):
    host = os.getenv("CLICKHOUSE_HOST", "127.0.0.1")
    port = int(os.getenv("CLICKHOUSE_TCP_PORT", "9000"))
    client = Client(host=host, port=port, database="analytics")

    batch = []
    inserted = 0
    for row in generate_events(total):
        batch.append(row)
        if len(batch) >= batch_size:
            client.execute(
                "INSERT INTO page_events (id, user_id, event_type, url, referrer, ts, meta) VALUES",
                batch
            )
            inserted += len(batch)
            print(f"Inserted {inserted} rows…")
            batch.clear()
    if batch:
        client.execute(
            "INSERT INTO page_events (id, user_id, event_type, url, referrer, ts, meta) VALUES",
            batch
        )
        inserted += len(batch)
        print(f"Inserted {inserted} rows (done)")

if __name__ == "__main__":
    total = int(sys.argv[1]) if len(sys.argv) > 1 else 1000000
    seed(total)

