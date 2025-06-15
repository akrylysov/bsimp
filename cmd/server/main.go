package main

import (
	"github.com/ginqi7/bsimp/internal/auth"
	"github.com/ginqi7/bsimp/internal/config"
	"github.com/ginqi7/bsimp/internal/media"
	"github.com/ginqi7/bsimp/internal/server"
	"github.com/ginqi7/bsimp/internal/storage"
)

import (
	"flag"
	"log/slog"
)

func main() {
	var (
		httpAddr   string
		configPath string
	)
	flag.StringVar(&httpAddr, "http", ":8080", "HTTP server address")
	flag.StringVar(&configPath, "config", "config.toml", "config path")
	flag.Parse()

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		slog.Error("failed parsing confg",
			slog.String("error", err.Error()),
			slog.String("path", configPath))
		return
	}

	universalStorage := storage.UniversalStorage{}

	if *cfg.S3.Type == "local" {
		localStore, err := storage.NewLocalStorage(cfg.S3)
		if err != nil {
			slog.Error("failed initializing S3 storage",
				slog.String("error", err.Error()))
			return
		}
		universalStorage.Local = localStore
	}

	if *cfg.S3.Type == "s3" {
		s3Store, err := storage.NewS3Storage(cfg.S3)
		if err != nil {
			slog.Error("failed initializing S3 storage",
				slog.String("error", err.Error()))
			return
		}
		universalStorage.S3 = s3Store
	}

	mediaLib := media.NewMediaLibrary(&universalStorage)
	authLib := auth.NewAuthLibrary(cfg)
	slog.Info("started HTTP server", slog.String("address", httpAddr))
	err = server.StartServer(mediaLib, authLib, httpAddr)
	slog.Error("failed starting HTTP server",
		slog.String("error", err.Error()))
}
