package main

import (
	"log"
	"net/http"

	"avatar/backend/internal/config"
	"avatar/backend/internal/db"
	"avatar/backend/internal/httpapi"
	"avatar/backend/internal/repo"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DSN)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	handler := httpapi.NewServer(repo.NewMySQLRepo(database))

	log.Printf("listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, handler); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
