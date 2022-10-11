package main

import (
	"errors"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Duration time.Duration

func (d *Duration) UnmarshalText(data []byte) error {
	val, err := time.ParseDuration(string(data))
	*d = Duration(val)
	return err
}

type S3Credentials struct {
	Id     string
	Secret string
	Token  string
}

type S3Config struct {
	Region               *string
	Endpoint             *string
	Bucket               string
	BasePrefix           string   `toml:"base_prefix"`
	RequestPresignExpiry Duration `toml:"request_presign_expiry"`
	Credentials          *S3Credentials
}

type Config struct {
	S3 S3Config
}

func NewConfig(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	if err := toml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}
	if cfg.S3.Bucket == "" {
		return nil, errors.New("s3 bucket is required")
	}
	return cfg, nil
}
