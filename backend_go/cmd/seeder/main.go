package main

import (
    "flag"
    "log"
    "project/pkg/seeder"
)

func main() {
    rows := flag.Int("rows", 1000000, "Number of rows to generate")
    batch := flag.Int("batch", 10000, "Batch size for loading")
    flag.Parse()

    err := seeder.Seed(*rows, *batch)
    if err != nil {
        log.Fatalf("Seeding failed: %v", err)
    }
    log.Printf("Seeded %d rows.", *rows)
}
