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
	BindAddr                   string `env:"BIND_ADDR"                        flag:"bind-addr"                      flagDesc:"Bind address"`
	MongoDBURL                 string `env:"MONGODB_URL"                      flag:"mongodb-url"                    flagDesc:"MongoDB server URL"`
	Database                   string `env:"INSOLVENCY_MONGODB_DATABASE"      flag:"mongodb-database"               flagDesc:"MongoDB database for data"`
	MongoCollection            string `env:"INSOLVENCY_MONGODB_COLLECTION"    flag:"mongodb-collection"             flagDesc:"The name of the mongodb collection"`
	IsEfsAllowListAuthDisabled bool   `env:"DISABLE_EFS_ALLOW_LIST_AUTH"      flag:"disable-efs-allow-list-auth"    flagDesc:"Set to 'true' in order to bypass EFS allow list aspect of API authorisation"`
	EnableNonLiveRouteHandlers bool   `env:"ENABLE_NON_LIVE_ROUTE_HANDLERS"     flag:"enable-non-live-route-handlers"   flagdesc:"Set to 'true'/'false' to respectively enable/disable form endpoints internal/external availability"`
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
