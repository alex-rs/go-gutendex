package internal

import (
	"context"
	"net/http"
	"time"

	httpcache "github.com/gregjones/httpcache"
	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

// Client is a thin wrapper providing caching, retry and rate limiting.
type Client struct {
	client  *retryablehttp.Client
	std     *http.Client
	Limiter *rate.Limiter
}

// New constructs a configured Client.
func New() *Client {
	cache := httpcache.NewTransport(httpcache.NewMemoryCache())
	rc := retryablehttp.NewClient()
	rc.RetryMax = 4
	rc.Backoff = retryablehttp.LinearJitterBackoff
	rc.HTTPClient.Transport = cache
	rc.Logger = nil
	rc.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if err != nil {
			return true, err
		}
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
			return true, nil
		}
		return false, nil
	}
	return &Client{
		client:  rc,
		std:     rc.StandardClient(),
		Limiter: rate.NewLimiter(rate.Every(time.Second), 1),
	}
}

// SetRetryWait overrides the retry backoff bounds ensuring min <= max.
func (c *Client) SetRetryWait(minDur, maxDur time.Duration) {
	c.client.RetryWaitMin = min(minDur, maxDur)
	c.client.RetryWaitMax = max(minDur, maxDur)
}

// SetRetryMax sets the maximum number of retries.
func (c *Client) SetRetryMax(max int) { c.client.RetryMax = max }

// SetCheckRetry overrides the retry check function.
func (c *Client) SetCheckRetry(fn retryablehttp.CheckRetry) { c.client.CheckRetry = fn }

// Do executes the HTTP request respecting rate limiting and retries.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if err := c.Limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return c.std.Do(req.WithContext(ctx))
}
