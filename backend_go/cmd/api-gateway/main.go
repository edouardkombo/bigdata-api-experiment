package main

import (
    "log"
    "net/http"
    "project/pkg/api"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/go-chi/cors"
)

func main() {
    r := chi.NewRouter()

    // Standard middlewares
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // CORS — allow your frontend origin (or "*" for all)
    r.Use(cors.Handler(cors.Options{
        // Allow requests from anywhere:
        AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
        // Or restrict to your domain:
        // AllowedOrigins:   []string{"https://seo-tools.pinnacle.com"},
        AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
        AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
        ExposedHeaders:   []string{"Link"},
        AllowCredentials: false,
        MaxAge:           300, // 5 minutes
    }))

    // API routes
    r.Get("/metrics/overview", api.OverviewHandler)
    r.Get("/metrics/events",   api.EventsHandler)
    r.Get("/metrics/time-series",   api.TimeSeriesHandler)
    r.Get("/metrics/type-breakdown", api.TypeBreakdownHandler)

    addr := ":8080"
    log.Printf("API Gateway listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, r))
}

