package discord

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

type ImageFetcher interface {
	Fetch(ctx context.Context, url string) (io.ReadCloser, error)
}

type httpImageFetcher struct {
	doer doer
}

func (f httpImageFetcher) Fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := f.doer.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, &httpError{StatusCode: resp.StatusCode, URL: url}
	}

	return resp.Body, nil
}

// httpError represents an HTTP error during image fetching.
type httpError struct {
	StatusCode int
	URL        string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("failed to fetch image from %s: status %d", e.URL, e.StatusCode)
}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}
