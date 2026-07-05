package main

import (
	"log"

	"quota/internal/config"
	"quota/internal/db"
	"quota/internal/router"
	"quota/internal/seed"
)

func main() {
	cfg := config.Load()

	gdb, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	if err := db.Migrate(gdb); err != nil {
		log.Fatalf("db migrate: %v", err)
	}

	// Seed demo data on first run so the app is explorable immediately.
	if err := seed.Run(gdb); err != nil {
		log.Printf("seed: %v", err)
	}

	r := router.New(gdb, cfg)
	log.Printf("listening on :%s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server: %v", err)
	}
}
