package internal

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestSetRetryWaitOrder(t *testing.T) {
	c := New()
	c.SetRetryWait(2*time.Second, 500*time.Millisecond)
	if c.client.RetryWaitMin > c.client.RetryWaitMax {
		t.Fatalf("min > max: %v > %v", c.client.RetryWaitMin, c.client.RetryWaitMax)
	}
	if c.client.RetryWaitMin != 500*time.Millisecond || c.client.RetryWaitMax != 2*time.Second {
		t.Fatalf("unexpected values: %v %v", c.client.RetryWaitMin, c.client.RetryWaitMax)
	}
}

func TestCheckRetryNetworkError(t *testing.T) {
	c := New()
	if retry, err := c.client.CheckRetry(context.Background(), nil, errors.New("network")); !retry || err != nil {
		t.Fatalf("expected retry with nil error, got %v, %v", retry, err)
	}
	if retry, err := c.client.CheckRetry(context.Background(), nil, context.Canceled); retry || !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got retry %v err %v", retry, err)
	}
}

type roundTripper func(*http.Request) (*http.Response, error)

func (rt roundTripper) RoundTrip(r *http.Request) (*http.Response, error) { return rt(r) }

func TestCheckRetryStatusCodes(t *testing.T) {
	c := New()
	tests := []struct {
		code      int
		wantRetry bool
	}{
		{http.StatusOK, false},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
	}
	for _, tt := range tests {
		resp := &http.Response{StatusCode: tt.code}
		retry, err := c.client.CheckRetry(context.Background(), resp, nil)
		if err != nil {
			t.Fatalf("unexpected error for code %d: %v", tt.code, err)
		}
		if retry != tt.wantRetry {
			t.Errorf("code %d: retry=%v want %v", tt.code, retry, tt.wantRetry)
		}
	}
}

func TestDoRetriesNetworkError(t *testing.T) {
	c := New()
	c.SetRetryWait(0, 0)
	attempts := 0
	rt := roundTripper(func(req *http.Request) (*http.Response, error) {
		attempts++
		if attempts < 3 {
			return nil, errors.New("temporary")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Header:     make(http.Header),
		}, nil
	})
	c.client.HTTPClient.Transport = rt
	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error creating request: %v", err)
	}
	resp, err := c.Do(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}
