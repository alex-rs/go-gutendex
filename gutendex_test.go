package gutendex

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	internal "github.com/alex-rs/go-gutendex/internal"
	"golang.org/x/time/rate"
	"io"
)

func newTestClient(baseURL string) *Client {
	c := &Client{hc: internal.New(), baseURL: baseURL}
	c.hc.Limiter = rate.NewLimiter(rate.Inf, 1)
	c.hc.SetRetryWait(0, 0)
	return c
}

func TestIteratorExhaustion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `{"count":1,"next":null,"previous":null,"results":[{"id":1,"title":"Test"}]}`)
	}))
	defer srv.Close()

	it := NewIter[Book](internal.New(), srv.URL)
	it.client.Limiter = rate.NewLimiter(rate.Inf, 1)
	it.client.SetRetryWait(0, 0)

	if !it.Next() {
		t.Fatalf("expected Next true")
	}
	if b := it.Value(); b.ID != 1 {
		t.Fatalf("unexpected book: %+v", b)
	}
	if it.Next() {
		t.Fatalf("iterator should be exhausted")
	}
	if err := it.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotFound(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := c.GetBook(context.Background(), 1)
	if err == nil || !IsNotFound(err) {
		t.Fatalf("expected IsNotFound, got %v", err)
	}
}

func TestRetryAfter429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = fmt.Fprint(w, `{"id":1,"title":"ok","authors":[],"translators":[],"subjects":[],"bookshelves":[],"languages":[],"copyright":null,"media_type":"","formats":{},"download_count":0}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	if _, err := c.GetBook(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestCaching(t *testing.T) {
	hits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Cache-Control", "max-age=60")
		_, _ = fmt.Fprint(w, `OK`)
	}))
	defer srv.Close()

	hc := internal.New()
	hc.Limiter = rate.NewLimiter(rate.Inf, 1)

	req1, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp1, err := hc.Do(context.Background(), req1)
	if err != nil {
		t.Fatalf("first request: %v", err)
	}
	if _, err := io.ReadAll(resp1.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := resp1.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}
	if resp1.Header.Get("X-From-Cache") != "" {
		t.Fatalf("first request should not be cached")
	}

	req2, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	resp2, err := hc.Do(context.Background(), req2)
	if err != nil {
		t.Fatalf("second request: %v", err)
	}
	if _, err := io.ReadAll(resp2.Body); err != nil {
		t.Fatalf("read body2: %v", err)
	}
	if err := resp2.Body.Close(); err != nil {
		t.Fatalf("close body2: %v", err)
	}
	if resp2.Header.Get("X-From-Cache") != "1" {
		t.Fatalf("expected X-From-Cache header, got %q", resp2.Header.Get("X-From-Cache"))
	}
	if hits != 1 {
		t.Fatalf("server should be hit once, got %d", hits)
	}
}

func TestGetJSONErrorKinds(t *testing.T) {
	t.Run("network", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		srv.Close()
		c := newTestClient(srv.URL)
		c.hc.SetRetryMax(0)
		c.hc.SetCheckRetry(func(context.Context, *http.Response, error) (bool, error) { return false, nil })
		_, err := c.GetBook(context.Background(), 1)
		if err == nil {
			t.Fatalf("expected error")
		}
		var e *Error
		if !errors.As(err, &e) || e.Kind != ErrNetwork {
			t.Fatalf("expected ErrNetwork, got %v", err)
		}
	})

	t.Run("server", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		c := newTestClient(srv.URL)
		c.hc.SetRetryMax(0)
		c.hc.SetCheckRetry(func(context.Context, *http.Response, error) (bool, error) { return false, nil })
		_, err := c.GetBook(context.Background(), 1)
		var e *Error
		if !errors.As(err, &e) || e.Kind != ErrServer {
			t.Fatalf("expected ErrServer, got %v", err)
		}
	})

	t.Run("rate limited", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()
		c := newTestClient(srv.URL)
		c.hc.SetRetryMax(0)
		c.hc.SetCheckRetry(func(context.Context, *http.Response, error) (bool, error) { return false, nil })
		_, err := c.GetBook(context.Background(), 1)
		var e *Error
		if !errors.As(err, &e) || e.Kind != ErrRateLimited {
			t.Fatalf("expected ErrRateLimited, got %v", err)
		}
	})
}

func TestIterFetchErrors(t *testing.T) {
	tests := []struct {
		status int
		kind   ErrorKind
	}{
		{http.StatusNotFound, ErrNotFound},
		{http.StatusInternalServerError, ErrServer},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d", tt.status), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer srv.Close()

			it := NewIter[Book](internal.New(), srv.URL)
			it.client.Limiter = rate.NewLimiter(rate.Inf, 1)
			it.client.SetRetryWait(0, 0)
			it.client.SetRetryMax(0)
			it.client.SetCheckRetry(func(context.Context, *http.Response, error) (bool, error) { return false, nil })

			err := it.fetch(context.Background())
			if e, ok := err.(*Error); !ok || e.Kind != tt.kind {
				t.Fatalf("expected kind %v, got %v", tt.kind, err)
			}
		})
	}
}

func TestValueBeforeNextPanics(t *testing.T) {
	t.Run("panic", func(t *testing.T) {
		it := NewIter[Book](internal.New(), "")
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic")
			}
		}()
		_ = it.Value()
	}) // expect panic
}
