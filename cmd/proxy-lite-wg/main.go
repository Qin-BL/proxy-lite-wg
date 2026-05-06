package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"proxy-lite-wg/internal/api"
	"proxy-lite-wg/internal/config"
	"proxy-lite-wg/internal/service"
	"proxy-lite-wg/internal/store"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := os.MkdirAll(cfg.DataDir(), 0o755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.XrayConfigPath), 0o755); err != nil {
		log.Fatalf("create runtime dir: %v", err)
	}

	repo, err := store.NewSQLiteRepository(cfg.DBPath)
	if err != nil {
		log.Fatalf("open repository: %v", err)
	}
	defer repo.Close()

	if err := repo.Migrate(); err != nil {
		log.Fatalf("migrate database: %v", err)
	}

	svc, err := service.NewXrayService(cfg, repo)
	if err != nil {
		log.Fatalf("build service: %v", err)
	}
	defer svc.Close()

	if err := svc.Initialize(context.Background()); err != nil {
		log.Fatalf("initialize runtime: %v", err)
	}

	handler, err := api.NewServer(cfg, svc)
	if err != nil {
		log.Fatalf("build api server: %v", err)
	}

	log.Printf("proxy-lite-tls listening on %s", cfg.AppAddr)
	if err := http.ListenAndServe(cfg.AppAddr, handler); err != nil {
		log.Fatalf("serve http: %v", err)
	}
}
