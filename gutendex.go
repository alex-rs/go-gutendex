package gutendex

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	internal "github.com/alex-rs/go-gutendex/internal"
)

// Person represents an individual in Gutendex responses.
type Person struct {
	Name      string `json:"name"`
	BirthYear *int   `json:"birth_year"`
	DeathYear *int   `json:"death_year"`
}

// Book models a Gutendex book resource.
type Book struct {
	ID            int               `json:"id"`
	Title         string            `json:"title"`
	Authors       []Person          `json:"authors"`
	Translators   []Person          `json:"translators"`
	Subjects      []string          `json:"subjects"`
	Bookshelves   []string          `json:"bookshelves"`
	Languages     []string          `json:"languages"`
	Copyright     *bool             `json:"copyright"`
	MediaType     string            `json:"media_type"`
	Formats       map[string]string `json:"formats"`
	DownloadCount int               `json:"download_count"`
}

// Client provides access to the Gutendex API.
type Client struct {
	hc      *internal.Client
	baseURL string
}

// NewClient constructs a new API client.
func NewClient() *Client {
	return &Client{
		hc:      internal.New(),
		baseURL: "https://gutendex.com",
	}
}

// ListBooks returns an iterator over books matching the query.
func (c *Client) ListBooks(q Query) *Iter[Book] {
	u, _ := url.Parse(c.baseURL + "/books")
	vals := q.Values()
	if len(vals) > 0 {
		u.RawQuery = vals.Encode()
	}
	return NewIter[Book](c.hc, u.String())
}

// GetBook retrieves a single book by ID.
func (c *Client) GetBook(ctx context.Context, id int) (*Book, error) {
	var b Book
	url := fmt.Sprintf("%s/books/%d", c.baseURL, id)
	if err := c.getJSON(ctx, url, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

// Search is a convenience wrapper performing a keyword author search.
func (c *Client) Search(keyword string) *Iter[Book] {
	return c.ListBooks(Query{Author: keyword})
}

func (c *Client) getJSON(ctx context.Context, url string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return &Error{Op: "getJSON", Kind: ErrNetwork, Err: err}
	}
	resp, err := c.hc.Do(ctx, req)
	if err != nil {
		return &Error{Op: "getJSON", Kind: ErrNetwork, Err: err}
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
		return &Error{Op: "getJSON", Kind: kind, Err: fmt.Errorf("status %d", resp.StatusCode)}
	}
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		return &Error{Op: "getJSON", Kind: ErrServer, Err: err}
	}
	return nil
}
