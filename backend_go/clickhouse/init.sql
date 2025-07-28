CREATE DATABASE IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.page_events (
  id String,
  user_id String,
  event_type String,
  url String,
  referrer String,
  ts DateTime,
  meta String
) ENGINE = MergeTree()
ORDER BY ts;

