package api

import (
    "encoding/json"
    "net/http"
    "strconv"
    "project/pkg/db"
)

func EventsHandler(w http.ResponseWriter, r *http.Request) {
    cursor := r.URL.Query().Get("cursor")
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit <= 0 {
        limit = 50
    }

    events, err := db.StreamEvents(cursor, limit)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(events)
}

