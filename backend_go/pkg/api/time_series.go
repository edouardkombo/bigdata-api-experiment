// backend_go/pkg/api/time_series.go
package api

import (
    "encoding/json"
    "net/http"
    "project/pkg/db"
)

func TimeSeriesHandler(w http.ResponseWriter, r *http.Request) {
    from     := r.URL.Query().Get("from")
    to       := r.URL.Query().Get("to")
    interval := r.URL.Query().Get("interval")  // e.g. "5 minute"

    data, err := db.GetTimeSeries(from, to, interval)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}

