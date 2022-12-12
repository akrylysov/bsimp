package main

import (
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

const Delimiter = "/"

type Duration time.Duration

func (d *Duration) UnmarshalText(data []byte) error {
	val, err := time.ParseDuration(string(data))
	*d = Duration(val)
	return err
}

type S3Credentials struct {
	ID     string
	Secret string
	Token  string
}

type S3Config struct {
	Region               *string
	Endpoint             *string
	Bucket               string
	BasePrefix           string   `toml:"base_prefix"`
	RequestPresignExpiry Duration `toml:"request_presign_expiry"`
	ForcePathStyle       bool     `toml:"force_path_style"`
	Credentials          *S3Credentials
}

type Config struct {
	S3 S3Config
}

var errMissingBucket = errors.New("s3 bucket is required")

func newConfig(r io.Reader) (*Config, error) {
	cfg := &Config{
		S3: S3Config{
			RequestPresignExpiry: Duration(2 * time.Hour),
		},
	}
	dec := toml.NewDecoder(r)
	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}
	if cfg.S3.Bucket == "" {
		return nil, errMissingBucket
	}
	if cfg.S3.BasePrefix != "" && !strings.HasSuffix(cfg.S3.BasePrefix, Delimiter) {
		cfg.S3.BasePrefix += Delimiter
	}
	return cfg, nil
}

func NewConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return newConfig(f)
}
