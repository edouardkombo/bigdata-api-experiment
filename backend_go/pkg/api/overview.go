package api

import (
    "encoding/json"
    "net/http"
    "project/pkg/db"
)

func OverviewHandler(w http.ResponseWriter, r *http.Request) {
    ov, err := db.GetOverview()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(ov)
}

