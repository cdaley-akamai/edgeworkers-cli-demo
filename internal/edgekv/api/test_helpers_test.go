package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/edgegrid"
	"github.com/akamai/AkamaiOPEN-edgegrid-golang/v13/pkg/session"
	"github.com/akamai/edgeworkers-cli/internal/apiclient"
	"github.com/stretchr/testify/require"
)

type testTransport struct {
	server *httptest.Server
}

func (transport *testTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	cloned := request.Clone(request.Context())
	cloned.URL.Scheme = "http"
	cloned.URL.Host = transport.server.Listener.Addr().String()
	return http.DefaultTransport.RoundTrip(cloned)
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	config := &edgegrid.Config{
		Host:         server.Listener.Addr().String(),
		ClientToken:  "client-token",
		ClientSecret: "client-secret",
		AccessToken:  "access-token",
	}
	sessionClient, err := session.New(
		session.WithSigner(config),
		session.WithClient(&http.Client{Transport: &testTransport{server: server}}),
	)
	require.NoError(t, err)
	return New(apiclient.FromSession(sessionClient))
}
