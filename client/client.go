package client

import (
	"errors"
	"net/http"
	"reflect"

	choauth2 "github.com/companieshouse/api-sdk-go/oauth2"
	"github.com/companieshouse/chs.go/log"
	goauth2 "golang.org/x/oauth2"

	"github.com/companieshouse/api-sdk-go/apikey"
	"github.com/companieshouse/go-session-handler/httpsession"
	"github.com/companieshouse/go-session-handler/session"
	"github.com/companieshouse/insolvency-api/config"

	publicSDK "github.com/companieshouse/api-sdk-go/companieshouseapi"
	privateSDK "github.com/companieshouse/private-api-sdk-go/companieshouseapi"
)

// GetSDK will return an instance of the Public Go SDK using an oauth2 authenticated
// HTTP client if possible, else an API-key authenticated HTTP client will be used
func GetSDK(req *http.Request) (*publicSDK.Service, error) {

	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	// override the main api and payments api if the cfg values have been set
	if len(cfg.ApiUrl) > 0 {
		publicSDK.BasePath = cfg.ApiUrl
	}

	if len(cfg.ApiUrl) > 0 {
		publicSDK.PaymentsBasePath = cfg.ApiUrl
	}

	if len(cfg.ApiUrl) > 0 {
		publicSDK.FileTransferBasePath = cfg.ApiUrl
	}

	if len(cfg.ApiKey) > 0 {
		publicSDK.FileTransferApiKey = cfg.ApiKey
	}

	hc, err := getHTTPClient(req)
	if err != nil {
		return nil, err
	}

	// Return an instance of the SDK
	return publicSDK.New(hc)
}

// GetPrivateSDK will return an instance of the Private Go SDK using an oauth2 authenticated
// HTTP client if possible, else an API-key authenticated HTTP client will be used
func GetPrivateSDK(req *http.Request) (*privateSDK.Service, error) {
	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	// override the main api and payments api if the cfg values have been set
	if len(cfg.ApiUrl) > 0 {
		privateSDK.BasePath = cfg.ApiUrl
	}

	if len(cfg.ApiUrl) > 0 {
		privateSDK.PaymentsBasePath = cfg.ApiUrl
	}

	if len(cfg.ApiUrl) > 0 {
		privateSDK.InternalAPIBasePath = cfg.ApiUrl
	}

	if len(cfg.ApiUrl) > 0 {
		privateSDK.AlphaKeyBasePath = cfg.ApiUrl
	}

	hc, err := getHTTPClient(req)
	if err != nil {
		return nil, err
	}

	// Return an instance of the SDK
	return privateSDK.New(hc)
}

// getHttpClient returns an Http Client. It will be either Oauth2 or API-key
// authenticated depending on whether an Oauth token can be procured from the
// session data on the request context
func getHTTPClient(req *http.Request) (*http.Client, error) {

	sess := httpsession.GetSessionFromRequest(req)

	var tok *goauth2.Token
	// Firstly, we'll try to derive an Oauth token from the sess data
	if sess != nil {
		tok = sess.GetOauth2Token()
	}

	var hc *http.Client
	var err error

	// Check the token exists because we prefer oauth
	if tok != nil {
		// If it exists, we'll use it to return an authenticated HTTP client
		hc, err = getOauth2HTTPClient(req, tok, sess)
	} else {
		// Otherwise, we'll use API-key authentication
		hc, err = getAPIKeyHTTPClient(req)
	}

	if err != nil {
		return nil, err
	}

	// Set the redirect suppression function - this is important because in the
	// case of the API returning a redirect, we don't necessarily want our
	// authenticated HTTP client following the redirection
	hc.CheckRedirect = deferRedirect

	return hc, nil
}

// getOauth2HttpClient returns an Oauth2-authenticated HTTP client
func getOauth2HTTPClient(req *http.Request, tok *goauth2.Token, s *session.Session) (*http.Client, error) {

	// Fetch oauth config
	oauth2Config, err := config.GetOauthConfig()
	if err != nil {
		return nil, err
	}

	// Initialise the callback function to be fired on session expiry
	var fn choauth2.NotifyFunc = AccessTokenChangedCallback

	// Create an http client
	return oauth2Config.Client(req.Context(), tok, fn, s), nil
}

// getAPIKeyHttpClient returns an API-key-authenticated HTTP client
func getAPIKeyHTTPClient(req *http.Request) (*http.Client, error) {

	cfg, err := config.Get()
	if err != nil {
		return nil, err
	}

	// Initialise an apikey cfg struct
	apiKeyConfig := &apikey.Config{Key: cfg.ApiKey}

	// Create an http client
	return apiKeyConfig.Client(req.Context(), cfg.ApiKey), nil
}

// deferRedirect will manually prevent the client from following a redirect;
// Instead, we will check for a redirect status on the response and take action
// accordingly.
// This is important in cases where we don't want to follow a redirect using an
// oauth authenticated client. For example, we don't want to pass bearer tokens
// when communicating with S3
func deferRedirect(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

// AccessTokenChangedCallback will be called when attempting to make an API call
// from an expired session. This function will refresh the access token on the
// session
func AccessTokenChangedCallback(newToken *goauth2.Token, private interface{}) error {
	sess, ok := (private).(*session.Session)
	if !ok {
		log.Error(errors.New("error converting private data to Session"),
			log.Data{
				"private_data_interface_type": reflect.TypeOf(private).String(),
			})
		return errors.New("error converting interface to Session")
	}

	sess.SetAccessToken(newToken.AccessToken)

	return sess.RefreshExpiration()
}
