package merkle

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTreeFetcherHostAllowlist(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tree file"))
	}))
	defer srv.Close()

	srvURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	f := NewHTTPTreeFetcher(srv.Client(), " merkle.dimo.zone, "+srvURL.Host)
	assert.Equal(t, []string{"merkle.dimo.zone", srvURL.Host}, f.AllowedHosts)

	// Allowed host works.
	body, err := f.Fetch(ctx, srv.URL+"/pool-0/week-214.json")
	require.NoError(t, err)
	assert.Equal(t, []byte("tree file"), body)

	// Hosts not in the list are rejected.
	_, err = f.Fetch(ctx, "https://evil.example.com/pool-0/week-214.json")
	assert.ErrorContains(t, err, "not in the allowed list")

	// Non-HTTPS schemes are rejected.
	_, err = f.Fetch(ctx, "http://"+srvURL.Host+"/pool-0/week-214.json")
	assert.ErrorContains(t, err, "does not use https")
}

func TestHTTPTreeFetcherRefusesRedirectToDisallowedHost(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://evil.example.com/pool-0/week-214.json", http.StatusFound)
	}))
	defer srv.Close()

	srvURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	client := srv.Client()
	f := NewHTTPTreeFetcher(client, srvURL.Host)

	// The original client must not be mutated.
	assert.Nil(t, client.CheckRedirect)

	// The initial host is allowed, but the redirect target is not.
	_, err = f.Fetch(ctx, srv.URL+"/pool-0/week-214.json")
	require.Error(t, err)
	assert.ErrorContains(t, err, "not in the allowed list")
}

func TestHTTPTreeFetcherStatusAndSizeLimits(t *testing.T) {
	ctx := context.Background()

	big := bytes.Repeat([]byte{0}, maxTreeFileSize+1)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing.json":
			w.WriteHeader(http.StatusNotFound)
		case "/big.json":
			_, _ = w.Write(big)
		}
	}))
	defer srv.Close()

	srvURL, err := url.Parse(srv.URL)
	require.NoError(t, err)

	f := NewHTTPTreeFetcher(srv.Client(), srvURL.Host)

	_, err = f.Fetch(ctx, srv.URL+"/missing.json")
	assert.ErrorContains(t, err, "status code 404")

	_, err = f.Fetch(ctx, srv.URL+"/big.json")
	assert.ErrorContains(t, err, "exceeds the size limit")
}
