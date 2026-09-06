package dnsmint

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-acme/lego/v5/internal/tester"
	"github.com/go-acme/lego/v5/internal/tester/servermock"
	"github.com/stretchr/testify/require"
)

var envTest = tester.NewEnvTest(EnvAPIKey, EnvAPIURL)

func TestNewDNSProvider(t *testing.T) {
	testCases := []struct {
		desc     string
		envVars  map[string]string
		expected string
	}{
		{
			desc: "success",
			envVars: map[string]string{
				EnvAPIKey: "dnsm_key_secret",
			},
		},
		{
			desc: "success with custom API URL",
			envVars: map[string]string{
				EnvAPIKey: "dnsm_key_secret",
				EnvAPIURL: "https://staging.dnsmint.example/api",
			},
		},
		{
			desc: "missing API key",
			envVars: map[string]string{
				EnvAPIKey: "",
			},
			expected: "dnsmint: some credentials information are missing: DNSMINT_API_KEY",
		},
		{
			desc: "invalid API URL",
			envVars: map[string]string{
				EnvAPIKey: "dnsm_key_secret",
				EnvAPIURL: ":",
			},
			expected: `dnsmint: parse ":": missing protocol scheme`,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			defer envTest.RestoreEnv()

			envTest.ClearEnv()

			envTest.Apply(test.envVars)

			p, err := NewDNSProvider()

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestNewDNSProviderConfig(t *testing.T) {
	apiURL, _ := url.Parse(defaultAPIURL)

	testCases := []struct {
		desc     string
		config   *Config
		expected string
	}{
		{
			desc:   "success",
			config: &Config{APIKey: "dnsm_key_secret", APIURL: apiURL},
		},
		{
			desc:     "nil config",
			expected: "dnsmint: the configuration of the DNS provider is nil",
		},
		{
			desc:     "missing API key",
			config:   &Config{APIURL: apiURL},
			expected: "dnsmint: credentials missing",
		},
		{
			desc:     "missing API URL",
			config:   &Config{APIKey: "dnsm_key_secret"},
			expected: "dnsmint: the API URL is missing",
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			p, err := NewDNSProviderConfig(test.config)

			if test.expected == "" {
				require.NoError(t, err)
				require.NotNil(t, p)
				require.NotNil(t, p.config)
			} else {
				require.EqualError(t, err, test.expected)
			}
		})
	}
}

func TestDNSProvider_Present(t *testing.T) {
	// The FQDN keeps the trailing dot lego computes; the server strips it.
	provider := mockBuilder().
		Route("/httpreq/present",
			servermock.RawStringResponse(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`),
			servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)).
		Build(t)

	err := provider.Present(t.Context(), "domain", "token", "key")
	require.NoError(t, err)
}

func TestDNSProvider_Present_error(t *testing.T) {
	// What a key narrowed to a different hostname gets; the status must reach the caller.
	provider := mockBuilder().
		Route("/httpreq/present",
			servermock.RawStringResponse(`{"error":"This API key's \"dns01:write\" scope is limited to other.example.dev","code":"FORBIDDEN"}`).
				WithStatusCode(http.StatusForbidden)).
		Build(t)

	err := provider.Present(t.Context(), "domain", "token", "key")
	require.Error(t, err)
	require.ErrorContains(t, err, "403")
}

func TestDNSProvider_CleanUp(t *testing.T) {
	provider := mockBuilder().
		Route("/httpreq/cleanup",
			servermock.RawStringResponse(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`),
			servermock.CheckRequestJSONBody(`{"fqdn":"_acme-challenge.domain.","value":"LHDhK3oGRvkiefQnx7OOczTY5Tic_xZ6HcMOc_gmtoM"}`)).
		Build(t)

	err := provider.CleanUp(t.Context(), "domain", "token", "key")
	require.NoError(t, err)
}

func TestDNSProvider_CleanUp_error(t *testing.T) {
	provider := mockBuilder().
		Route("/httpreq/cleanup", servermock.Noop().WithStatusCode(http.StatusUnauthorized)).
		Build(t)

	err := provider.CleanUp(t.Context(), "domain", "token", "key")
	require.Error(t, err)
	require.ErrorContains(t, err, "401")
}

func mockBuilder() *servermock.Builder[*DNSProvider] {
	return servermock.NewBuilder(
		func(server *httptest.Server) (*DNSProvider, error) {
			config := NewDefaultConfig()
			config.HTTPClient = server.Client()
			config.APIKey = "dnsm_key_secret"
			config.APIURL, _ = url.Parse(server.URL)

			return NewDNSProviderConfig(config)
		},
		servermock.CheckHeader().
			WithJSONHeaders().
			WithAuthorization("Bearer dnsm_key_secret"),
	)
}
