package gutendex

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestQueryValues(t *testing.T) {
	tests := []struct {
		name string
		q    Query
		want url.Values
	}{
		{
			name: "author and title combine to search",
			q:    Query{Author: "a", Title: "t"},
			want: url.Values{"search": {"a t"}},
		},
		{
			name: "only author",
			q:    Query{Author: "a"},
			want: url.Values{"author": {"a"}},
		},
		{
			name: "only title",
			q:    Query{Title: "t"},
			want: url.Values{"title": {"t"}},
		},
		{
			name: "topic language mime",
			q:    Query{Topic: "top", Language: "en", MIME: "text"},
			want: url.Values{"topic": {"top"}, "languages": {"en"}, "mime_type": {"text"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.q.Values()
			if got.Encode() != tt.want.Encode() {
				t.Fatalf("Values() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestListBooksAndSearchURLs(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.String()
		_, _ = w.Write([]byte(`{"count":0,"next":null,"previous":null,"results":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)

	it := c.ListBooks(Query{Title: "foo"})
	if it.Next() {
		t.Fatalf("expected no results")
	}
	if got != "/books?title=foo" {
		t.Fatalf("ListBooks URL = %s", got)
	}

	it = c.Search("bar")
	if it.Next() {
		t.Fatalf("expected no results")
	}
	if got != "/books?author=bar" {
		t.Fatalf("Search URL = %s", got)
	}
}
