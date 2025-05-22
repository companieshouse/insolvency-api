// Package config defines the environment variable and command-line flags
package config

import (
	"sync"

	choauth2 "github.com/companieshouse/api-sdk-go/oauth2"
	"github.com/companieshouse/gofigure"
	goauth2 "golang.org/x/oauth2"
)

var cfg *Config
var mtx sync.Mutex
var oauthConfig *choauth2.Config

// Config defines the configuration options for this service.
type Config struct {
	ClientID                   string   `env:"OAUTH2_CLIENT_ID"                 flag:"oauth2-local-client-id"               flagDesc:"Client ID"`
	ClientSecret               string   `env:"OAUTH2_CLIENT_SECRET"             flag:"oauth2-local-client-secret"           flagDesc:"Client Secret"`
	RedirectURL                string   `env:"OAUTH2_REDIRECT_URI"              flag:"oauth2-local-redirect-uri"            flagDesc:"Oauth2 Redirect URI"`
	AuthURL                    string   `env:"OAUTH2_AUTH_URI"                  flag:"oauth2-local-auth-uri"                flagDesc:"Oauth2 Auth URI"`
	TokenURL                   string   `env:"OAUTH2_TOKEN_URI"                 flag:"oauth2-tlocal-oken-uri"               flagDesc:"Oauth2 Token URI"`
	Scopes                     []string `env:"SCOPE"                            flag:"local-scope"                          flagDesc:"Scope"`
	ApiUrl                     string   `env:"API_URL"                          flag:"api-local-url"                        flagDesc:"Api Url"`
	ApiKey                     string   `env:"API_KEY"                          flag:"api-local-key"                        flagDesc:"Api Key"`
	BindAddr                   string   `env:"BIND_ADDR"                        flag:"bind-addr"                      flagDesc:"Bind address"`
	MongoDBURL                 string   `env:"MONGODB_URL"                      flag:"mongodb-url"                    flagDesc:"MongoDB server URL"`
	Database                   string   `env:"INSOLVENCY_MONGODB_DATABASE"      flag:"mongodb-database"               flagDesc:"MongoDB database for data"`
	MongoCollection            string   `env:"INSOLVENCY_MONGODB_COLLECTION"    flag:"mongodb-collection"             flagDesc:"The name of the mongodb collection"`
	IsEfsAllowListAuthDisabled bool     `env:"DISABLE_EFS_ALLOW_LIST_AUTH"      flag:"disable-efs-allow-list-auth"    flagDesc:"Set to 'true' in order to bypass EFS allow list aspect of API authorisation"`
	EnableNonLiveRouteHandlers bool     `env:"ENABLE_NON_LIVE_ROUTE_HANDLERS"     flag:"enable-non-live-route-handlers"   flagdesc:"Set to 'true'/'false' to respectively enable/disable form endpoints internal/external availability"`
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

// GetOauthConfig returns an instance of a Companies House oauth config struct
// and an error where appropriate
func GetOauthConfig() (*choauth2.Config, error) {

	if oauthConfig != nil {
		return oauthConfig, nil
	}

	config, err := Get()
	if err != nil {
		return nil, err
	}

	oauthConfig = &choauth2.Config{}
	oauthConfig.ClientID = config.ClientID
	oauthConfig.ClientSecret = config.ClientSecret
	oauthConfig.RedirectURL = config.RedirectURL
	oauthConfig.Scopes = config.Scopes
	oauthConfig.Endpoint = goauth2.Endpoint{
		AuthURL:  config.AuthURL,
		TokenURL: config.TokenURL,
	}

	return oauthConfig, nil
}
