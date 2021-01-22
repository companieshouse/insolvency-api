// Package config defines the environment variable and command-line flags
package config

import (
	"sync"

	"github.com/companieshouse/gofigure"
)

var cfg *Config
var mtx sync.Mutex

// Config defines the configuration options for this service.
type Config struct {
	BindAddr   string   `env:"BIND_ADDR"         flag:"bind-addr"   flagDesc:"Bind address"`
	CHSURL     string   `env:"CHS_URL"           flag:"chs-url"     flagDesc:"CHS URL"`
	BrokerAddr []string `env:"KAFKA_BROKER_ADDR" flag:"broker-addr" flagDesc:"Kafka broker address"`
}

// Get returns a pointer to a Config instance populated with values from environment or command-line flags
func Get() (*Config, error) {
	mtx.Lock()
	defer mtx.Unlock()

	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{}

	err := gofigure.Gofigure(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
