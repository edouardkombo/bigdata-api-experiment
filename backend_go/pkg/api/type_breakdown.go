package api

import (
    "encoding/json"
    "net/http"
    "project/pkg/db"
)

func TypeBreakdownHandler(w http.ResponseWriter, r *http.Request) {
    data, err := db.GetTypeBreakdown()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

