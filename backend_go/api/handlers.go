package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/ClickHouse/clickhouse-go"
)

var dsn = "tcp://127.0.0.1:9000?debug=false"

func setSessionLimits(db *sql.DB) {
	db.Exec(`
		SET max_memory_usage = 6000000000,
		    max_bytes_before_external_group_by = 100000000,
		    max_bytes_before_external_sort = 100000000,
		    max_threads = 4
	`)
}

func OverviewHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer db.Close()
	setSessionLimits(db)

	row := db.QueryRow(`
		SELECT count() AS total,
		       uniq(user_id) AS unique_users,
		       min(ts),
		       max(ts)
		FROM analytics.page_events
	`)
	var total int
	var users int
	var minTS, maxTS string
	if err := row.Scan(&total, &users, &minTS, &maxTS); err != nil {
		http.Error(w, "Scan error", 500)
		return
	}

	rows, err := db.Query(`
		SELECT id, user_id, event_type, url, referrer, ts, meta
		FROM analytics.page_events
		WHERE ts > now() - INTERVAL 1 HOUR
		ORDER BY ts DESC
		LIMIT 500000
	`)
	if err != nil {
		http.Error(w, "Query error", 500)
		return
	}
	defer rows.Close()

	recent := []map[string]any{}
	for rows.Next() {
		var id, userID, eventType, url, referrer, ts, meta string
		if err := rows.Scan(&id, &userID, &eventType, &url, &referrer, &ts, &meta); err != nil {
			continue
		}
		var metaMap map[string]string
		if err := json.Unmarshal([]byte(meta), &metaMap); err != nil {
			log.Printf("\u26a0\ufe0f Failed to decode meta for ID %s: %v", id, err)
			metaMap = map[string]string{"raw": meta}
		}

		e := map[string]any{
			"id":         id,
			"user_id":    userID,
			"event_type": eventType,
			"url":        url,
			"referrer":   referrer,
			"ts":         ts,
			"meta":       metaMap,
		}
		recent = append(recent, e)
	}

	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(map[string]any{
		"total":        total,
		"unique_users": users,
		"first_event":  minTS,
		"last_event":   maxTS,
		"last_hour":    recent,
	})
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer db.Close()
	setSessionLimits(db)

	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}

	userID := r.URL.Query().Get("user_id")
	eventType := r.URL.Query().Get("event_type")

	whereClauses := []string{}
	args := []any{}

	if userID != "" {
		whereClauses = append(whereClauses, "user_id = ?")
		args = append(args, userID)
	}
	if eventType != "" {
		whereClauses = append(whereClauses, "event_type = ?")
		args = append(args, eventType)
	}

	query := `SELECT id, user_id, event_type, url, referrer, ts, meta FROM analytics.page_events`
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	query += " ORDER BY ts DESC LIMIT ?"
	args = append(args, limit)

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), 500)
		return
	}
	defer rows.Close()

	events := []map[string]any{}
	for rows.Next() {
		var id, userID, eventType, url, referrer, ts, meta string
		if err := rows.Scan(&id, &userID, &eventType, &url, &referrer, &ts, &meta); err != nil {
			continue
		}
		var metaMap map[string]string
		json.Unmarshal([]byte(meta), &metaMap)
		e := map[string]any{
			"id":         id,
			"user_id":    userID,
			"event_type": eventType,
			"url":        url,
			"referrer":   referrer,
			"ts":         ts,
			"meta":       metaMap,
		}
		events = append(events, e)
	}
	json.NewEncoder(w).Encode(events)
}

func TimeSeriesHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		http.Error(w, "DB connection error", http.StatusInternalServerError)
		return
	}
	defer db.Close()
	setSessionLimits(db)

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	interval := r.URL.Query().Get("interval")
	if from == "" || to == "" || interval == "" {
		http.Error(w, "Missing query params: from, to, interval", http.StatusBadRequest)
		return
	}

	allowedIntervals := map[string]bool{
		"1 minute": true,
		"5 minute": true,
		"15 minute": true,
		"1 hour": true,
		"1 day": true,
	}
	if !allowedIntervals[interval] {
		http.Error(w, "Invalid interval", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf(`
		SELECT toStartOfInterval(ts, INTERVAL %s) AS bucket, count() as count
		FROM analytics.page_events
		WHERE ts BETWEEN parseDateTimeBestEffort(?) AND parseDateTimeBestEffort(?)
		GROUP BY bucket
		ORDER BY bucket ASC
		LIMIT 1000`, interval)

	rows, err := db.Query(query, from, to)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	points := []map[string]any{}
	for rows.Next() {
		var bucket string
		var count int
		if err := rows.Scan(&bucket, &count); err == nil {
			points = append(points, map[string]any{
				"bucket": bucket,
				"count":  count,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	json.NewEncoder(w).Encode(points)
}

func TypeBreakdownHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("clickhouse", dsn)
	if err != nil {
		http.Error(w, "DB error", 500)
		return
	}
	defer db.Close()
	setSessionLimits(db)

	rows, err := db.Query(`
		SELECT event_type, count() as c
		FROM analytics.page_events
		GROUP BY event_type
		ORDER BY c DESC
		LIMIT 100`)
	if err != nil {
		http.Error(w, "Query error", 500)
		return
	}
	defer rows.Close()

	typeCounts := make(map[string]int)
	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err == nil {
			typeCounts[eventType] = count
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(typeCounts)
}

