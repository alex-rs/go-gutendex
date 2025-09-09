package gutendex

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListBooksQueryURL(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.String()
		_, _ = fmt.Fprint(w, `{"count":0,"results":[]}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	it := c.ListBooks(Query{Title: "foo"})
	it.Next()

	if want := "/books?title=foo"; got != want {
		t.Fatalf("expected request to %q, got %q", want, got)
	}
}

func TestSearchDelegatesToListBooks(t *testing.T) {
	var got string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.String()
		_, _ = fmt.Fprint(w, `{"count":0,"results":[]}`)
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	it := c.Search("bar")
	it.Next()

	if want := "/books?author=bar"; got != want {
		t.Fatalf("expected request to %q, got %q", want, got)
	}
}
