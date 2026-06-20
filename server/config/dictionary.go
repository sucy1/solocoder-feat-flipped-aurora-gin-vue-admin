package config

import "time"

type Dictionary struct {
	TTL time.Duration `mapstructure:"ttl" json:"ttl" yaml:"ttl"`
}
