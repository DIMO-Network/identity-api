package merkle

import (
	"context"
	"errors"
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
// allowed hosts. The given client is copied and fitted with a redirect check
// that re-validates every redirect target against the allowlist, so redirects
// cannot be used to escape it.
func NewHTTPTreeFetcher(client *http.Client, allowedHosts string) *HTTPTreeFetcher {
	var hosts []string
	for h := range strings.SplitSeq(allowedHosts, ",") {
		if h = strings.TrimSpace(h); h != "" {
			hosts = append(hosts, strings.ToLower(h))
		}
	}

	f := &HTTPTreeFetcher{AllowedHosts: hosts}

	guarded := *client
	guarded.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return errors.New("stopped after 10 redirects")
		}
		return f.validateURL(req.URL)
	}
	f.Client = &guarded

	return f
}

// validateURL checks that the URL uses HTTPS and that its host is in the
// allowlist.
func (f *HTTPTreeFetcher) validateURL(u *url.URL) error {
	if u.Scheme != "https" {
		return fmt.Errorf("URL %q does not use https", u.Redacted())
	}
	if !slices.Contains(f.AllowedHosts, strings.ToLower(u.Host)) {
		return fmt.Errorf("host %q is not in the allowed list", u.Host)
	}
	return nil
}

// Fetch retrieves the file at the given URI, enforcing the host allowlist and
// the response size cap.
func (f *HTTPTreeFetcher) Fetch(ctx context.Context, uri string) ([]byte, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("parsing proofs URI %q: %w", uri, err)
	}
	if err := f.validateURL(u); err != nil {
		return nil, fmt.Errorf("proofs URI %q: %w", uri, err)
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
