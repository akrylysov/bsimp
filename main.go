package main

import (
	"flag"
	"log"
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
		log.Fatal(err)
	}

	store, err := NewS3Storage(cfg.S3)
	if err != nil {
		log.Fatal(err)
	}

	mediaLib := NewMediaLibrary(store)

	log.Fatal(StartServer(mediaLib, httpAddr))
}
