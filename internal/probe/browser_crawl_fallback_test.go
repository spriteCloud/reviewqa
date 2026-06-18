package probe

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/reviewqa/reviewqa/internal/mindmap"
	"github.com/reviewqa/reviewqa/internal/probe/browser"
)

// v0.88: when the browser runner can't start (node missing, npm
// install failed, no network), `--browser always` must surface the
// problem instead of silently degrading to static. A real crawl that
// just came back empty is a different signal — that one still falls
// back to static.

func TestCrawlOriginWithFallback_AlwaysBailsOnUnavailable(t *testing.T) {
	prev := browserCrawler
	t.Cleanup(func() { browserCrawler = prev })
	browserCrawler = func(ctx context.Context, u string) (*mindmap.Map, []error) {
		return nil, []error{fmt.Errorf("%w: node missing", browser.ErrBrowserUnavailable)}
	}

	var fetcherCalls atomic.Int32
	fetch := func(ctx context.Context, u string) ([]byte, string, error) {
		fetcherCalls.Add(1)
		return nil, "", errors.New("should not be called")
	}

	m, errs := crawlOriginWithFallback(context.Background(), "https://example.com/", mindmap.Fetcher(fetch), mindmap.Options{}, BrowserAlways)
	if m != nil {
		t.Errorf("expected nil map on Unavailable; got %+v", m)
	}
	if len(errs) == 0 || !errors.Is(errs[0], browser.ErrBrowserUnavailable) {
		t.Errorf("expected ErrBrowserUnavailable in errs; got %v", errs)
	}
	if got := fetcherCalls.Load(); got != 0 {
		t.Errorf("static fetcher called %d times; want 0 (BrowserAlways must not silently degrade)", got)
	}
}

func TestCrawlOriginWithFallback_AlwaysFallsBackOnZeroPages(t *testing.T) {
	prev := browserCrawler
	t.Cleanup(func() { browserCrawler = prev })
	browserCrawler = func(ctx context.Context, u string) (*mindmap.Map, []error) {
		// Empty mindmap with no Unavailable signal — a real browser
		// crawl that just found nothing. Static fallback is correct.
		return &mindmap.Map{Origin: u, Pages: map[string]*mindmap.Page{}}, nil
	}

	var fetcherCalls atomic.Int32
	fetch := func(ctx context.Context, u string) ([]byte, string, error) {
		fetcherCalls.Add(1)
		return []byte("<html><body>ok</body></html>"), "text/html", nil
	}

	_, _ = crawlOriginWithFallback(context.Background(), "https://example.com/", mindmap.Fetcher(fetch), mindmap.Options{}, BrowserAlways)

	if got := fetcherCalls.Load(); got == 0 {
		t.Error("expected static fetcher to be called on browser-returned-zero-pages; got 0 calls")
	}
}
