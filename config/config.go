// Package config defines the environment variable and command-line flags
package config

import (
	"github.com/companieshouse/gofigure"
	"sync"
)

var cfg *Config
var mtx sync.Mutex

type Config struct {
	BindAddr string   `env:"BIND_ADDR" flag:"bind-addr" flagDesc:"Bind address"`
  CHSURL   string   `env:"CHS_URL"   flag:"chs-url"   flagDesc:"CHS URL"`
}

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
