package main

import (
	"flag"
	"math/rand"
	"time"

	"golang.org/x/exp/slog"
)

func init() {
	// Prior to Go 1.20, the random generator is always seeded with Seed(1).
	rand.Seed(time.Now().UnixNano())
}

func main() {
	var (
		httpAddr   string
		configPath string
	)
	flag.StringVar(&httpAddr, "http", ":8080", "HTTP server address")
	flag.StringVar(&configPath, "config", "config.toml", "config path")
	flag.Parse()

	cfg, err := NewConfig(configPath)
	if err != nil {
		slog.Error("failed parsing confg", err, slog.String("path", configPath))
		return
	}

	store, err := NewS3Storage(cfg.S3)
	if err != nil {
		slog.Error("failed initializing S3 storage", err)
		return
	}

	mediaLib := NewMediaLibrary(store)

	slog.Info("started HTTP server", slog.String("address", httpAddr))
	err = StartServer(mediaLib, httpAddr)
	slog.Error("failed starting HTTP server", err)
}
