package merkle

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"
)

// maxTreeFileSize is the maximum number of bytes we are willing to read from a
// Merkle tree file referenced by a RootSet event.
const maxTreeFileSize = 50 * 1024 * 1024

// TreeFetcher retrieves the Merkle tree file referenced by a RootSet event's
// proofsURI.
type TreeFetcher interface {
	Fetch(ctx context.Context, uri string) ([]byte, error)
}

// HTTPTreeFetcher fetches tree files over HTTPS. Only hosts in AllowedHosts
// are permitted, and responses are capped at maxTreeFileSize bytes.
type HTTPTreeFetcher struct {
	Client       *http.Client
	AllowedHosts []string
}

// NewHTTPTreeFetcher creates an HTTPTreeFetcher from a comma-separated list of
// allowed hosts.
func NewHTTPTreeFetcher(client *http.Client, allowedHosts string) *HTTPTreeFetcher {
	var hosts []string
	for h := range strings.SplitSeq(allowedHosts, ",") {
		if h = strings.TrimSpace(h); h != "" {
			hosts = append(hosts, strings.ToLower(h))
		}
	}
	return &HTTPTreeFetcher{Client: client, AllowedHosts: hosts}
}

// Fetch retrieves the file at the given URI, enforcing the host allowlist and
// the response size cap.
func (f *HTTPTreeFetcher) Fetch(ctx context.Context, uri string) ([]byte, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("parsing proofs URI %q: %w", uri, err)
	}
	if u.Scheme != "https" {
		return nil, fmt.Errorf("proofs URI %q does not use https", uri)
	}
	if !slices.Contains(f.AllowedHosts, strings.ToLower(u.Host)) {
		return nil, fmt.Errorf("proofs URI host %q is not in the allowed list", u.Host)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request for %q: %w", uri, err)
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching %q: %w", uri, err)
	}
	defer resp.Body.Close() //nolint:errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %q: status code %d", uri, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxTreeFileSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading body of %q: %w", uri, err)
	}
	if len(body) > maxTreeFileSize {
		return nil, fmt.Errorf("tree file at %q exceeds the size limit of %d bytes", uri, maxTreeFileSize)
	}

	return body, nil
}
