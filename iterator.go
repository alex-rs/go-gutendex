package gutendex

import (
	"context"
	"encoding/json"
	"fmt"
	internal "github.com/alex-rs/go-gutendex/internal"
	"net/http"
)

// Page represents a paginated response from Gutendex.
type Page[T any] struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []T     `json:"results"`
}

// Iter iterates over items of type T from paginated endpoints.
type Iter[T any] struct {
	client  *internal.Client
	nextURL string
	buf     []T
	idx     int
	err     error
}

// NewIter constructs a new iterator starting at firstURL.
func NewIter[T any](client *internal.Client, firstURL string) *Iter[T] {
	return &Iter[T]{client: client, nextURL: firstURL, idx: -1}
}

// Next advances the iterator to the next value.
func (it *Iter[T]) Next() bool {
	if it.err != nil {
		return false
	}
	it.idx++
	if it.idx >= len(it.buf) {
		if it.nextURL == "" {
			return false
		}
		if err := it.fetch(context.Background()); err != nil {
			it.err = err
			return false
		}
		if len(it.buf) == 0 {
			return false
		}
		it.idx = 0
	}
	return true
}

// Value returns the current item. It panics if called without Next.
func (it *Iter[T]) Value() T {
	if it.idx < 0 || it.idx >= len(it.buf) {
		panic("iterator.Value called without Next")
	}
	return it.buf[it.idx]
}

// Err returns the last error encountered by the iterator.
func (it *Iter[T]) Err() error { return it.err }

func (it *Iter[T]) fetch(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, it.nextURL, nil)
	if err != nil {
		return &Error{Op: "iter.fetch", Kind: ErrNetwork, Err: err}
	}
	resp, err := it.client.Do(ctx, req)
	if err != nil {
		return &Error{Op: "iter.fetch", Kind: ErrNetwork, Err: err}
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		kind := ErrServer
		switch resp.StatusCode {
		case http.StatusNotFound:
			kind = ErrNotFound
		case http.StatusTooManyRequests:
			kind = ErrRateLimited
		}
		return &Error{Op: "iter.fetch", Kind: kind, Err: fmt.Errorf("status %d", resp.StatusCode)}
	}
	var page Page[T]
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return &Error{Op: "iter.fetch", Kind: ErrServer, Err: err}
	}
	it.buf = append(it.buf[:0], page.Results...)
	if page.Next != nil {
		it.nextURL = *page.Next
	} else {
		it.nextURL = ""
	}
	return nil
}
