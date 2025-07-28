package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"bigdata-perf/api"
)

func main() {
	envPath := filepath.Join("..", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("⚠️  Could not load .env from %s: %v", envPath, err)
	} else {
		log.Printf("✅ Loaded .env from %s", envPath)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8088"
		log.Println("⚠️  API_PORT not set, defaulting to 8088")
	} else {
		log.Printf("🔌 API_PORT from .env: %s", port)
	}

	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{"Link"},
		AllowCredentials: false,
		MaxAge: 300,
	}))
	r.Use(middleware.Logger)

	r.Get("/metrics/overview", api.OverviewHandler)
	r.Get("/metrics/events", api.EventsHandler)
	r.Get("/metrics/time-series", api.TimeSeriesHandler)
	r.Get("/metrics/type-breakdown", api.TypeBreakdownHandler)

	log.Printf("🚀 API running on :%s", port)
	http.ListenAndServe(":"+port, r)
}

