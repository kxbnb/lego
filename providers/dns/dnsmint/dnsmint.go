// Package dnsmint implements a DNS provider for solving the DNS-01 challenge using DNSMint.
package dnsmint

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/go-acme/lego/v5/challenge"
	"github.com/go-acme/lego/v5/challenge/dns01"
	"github.com/go-acme/lego/v5/internal/errutils"
	"github.com/go-acme/lego/v5/platform/env"
	"github.com/go-acme/lego/v5/providers/dns/internal/clientdebug"
)

// Environment variables names.
const (
	envNamespace = "DNSMINT_"

	EnvAPIKey = envNamespace + "API_KEY"
	EnvAPIURL = envNamespace + "API_URL"

	EnvPropagationTimeout = envNamespace + "PROPAGATION_TIMEOUT"
	EnvPollingInterval    = envNamespace + "POLLING_INTERVAL"
	EnvHTTPTimeout        = envNamespace + "HTTP_TIMEOUT"
)

const defaultAPIURL = "https://dnsmint.com/api"

var _ challenge.ProviderTimeout = (*DNSProvider)(nil)

type message struct {
	FQDN  string `json:"fqdn"`
	Value string `json:"value"`
}

// Config is used to configure the creation of the DNSProvider.
type Config struct {
	APIKey             string
	APIURL             *url.URL
	PropagationTimeout time.Duration
	PollingInterval    time.Duration
	HTTPClient         *http.Client
}

// NewDefaultConfig returns a default configuration for the DNSProvider.
func NewDefaultConfig() *Config {
	return &Config{
		PropagationTimeout: env.GetOrDefaultSecond(EnvPropagationTimeout, dns01.DefaultPropagationTimeout),
		PollingInterval:    env.GetOrDefaultSecond(EnvPollingInterval, dns01.DefaultPollingInterval),
		HTTPClient: &http.Client{
			Timeout: env.GetOrDefaultSecond(EnvHTTPTimeout, 30*time.Second),
		},
	}
}

// DNSProvider implements the challenge.Provider interface.
type DNSProvider struct {
	config *Config
}

// NewDNSProvider returns a DNSProvider instance configured for DNSMint.
// Credentials must be passed in the environment variable: DNSMINT_API_KEY.
func NewDNSProvider() (*DNSProvider, error) {
	values, err := env.Get(EnvAPIKey)
	if err != nil {
		return nil, fmt.Errorf("dnsmint: %w", err)
	}

	config := NewDefaultConfig()
	config.APIKey = values[EnvAPIKey]

	rawURL := env.GetOrDefaultString(EnvAPIURL, defaultAPIURL)

	config.APIURL, err = url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("dnsmint: %w", err)
	}

	return NewDNSProviderConfig(config)
}

// NewDNSProviderConfig return a DNSProvider instance configured for DNSMint.
func NewDNSProviderConfig(config *Config) (*DNSProvider, error) {
	if config == nil {
		return nil, errors.New("dnsmint: the configuration of the DNS provider is nil")
	}

	if config.APIKey == "" {
		return nil, errors.New("dnsmint: credentials missing")
	}

	if config.APIURL == nil {
		return nil, errors.New("dnsmint: the API URL is missing")
	}

	config.HTTPClient = clientdebug.Wrap(config.HTTPClient)

	return &DNSProvider{config: config}, nil
}

// Timeout returns the timeout and interval to use when checking for DNS propagation.
// Adjusting here to cope with spikes in propagation times.
func (d *DNSProvider) Timeout() (timeout, interval time.Duration) {
	return d.config.PropagationTimeout, d.config.PollingInterval
}

// Present creates a TXT record to fulfill the dns-01 challenge.
func (d *DNSProvider) Present(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	err := d.doPost(ctx, "/httpreq/present", info)
	if err != nil {
		return fmt.Errorf("dnsmint: %w", err)
	}

	return nil
}

// CleanUp removes the TXT record matching the specified parameters.
func (d *DNSProvider) CleanUp(ctx context.Context, domain, _, keyAuth string) error {
	info := dns01.GetChallengeInfo(ctx, domain, keyAuth)

	err := d.doPost(ctx, "/httpreq/cleanup", info)
	if err != nil {
		return fmt.Errorf("dnsmint: %w", err)
	}

	return nil
}

// The hostname is derived server-side from the challenge FQDN, so no zone lookup is needed.
// A key narrowed to a different hostname is refused with 403.
func (d *DNSProvider) doPost(ctx context.Context, uri string, info dns01.ChallengeInfo) error {
	reqBody := new(bytes.Buffer)

	err := json.NewEncoder(reqBody).Encode(message{FQDN: info.EffectiveFQDN, Value: info.Value})
	if err != nil {
		return fmt.Errorf("failed to create request JSON body: %w", err)
	}

	endpoint := d.config.APIURL.JoinPath(uri)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), reqBody)
	if err != nil {
		return fmt.Errorf("unable to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.config.APIKey)

	resp, err := d.config.HTTPClient.Do(req)
	if err != nil {
		return errutils.NewHTTPDoError(req, err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		return errutils.NewUnexpectedResponseStatusCodeError(req, resp)
	}

	return nil
}
