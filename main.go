package main

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

	cfg, err := NewConfig(configPath)
	if err != nil {
		slog.Error("failed parsing config", slog.Any("err", err), slog.String("path", configPath))
		return
	}

	mediaLib := NewMediaLibrary(NewS3Storage(cfg.S3))

	slog.Info("started HTTP server", slog.String("address", httpAddr))
	err = StartServer(mediaLib, httpAddr)
	slog.Error("failed starting HTTP server", slog.Any("err", err))
}
