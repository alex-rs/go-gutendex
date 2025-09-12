# Go Gutendex

[![Go Report Card](https://goreportcard.com/badge/github.com/alex-rs/go-gutendex)](https://goreportcard.com/report/github.com/alex-rs/go-gutendex)

Go Gutendex is a Go client for the [Gutendex](https://gutendex.com/) API, which exposes Project Gutenberg's catalog through a JSON interface. The package provides typed models, iterator-based pagination and convenience helpers for common queries.

## Features

- Minimal dependencies and idiomatic Go API
- Iterator abstraction for traversing paginated results
- Query helpers for filtering by author, title, topic, language or MIME type
- Simple method for fetching a book by its identifier

## Installation

```
go get github.com/alex-rs/go-gutendex
```

## Basic Usage

```go
package main

import (
    "context"
    "fmt"

    gutendex "github.com/alex-rs/go-gutendex"
)

func main() {
    client := gutendex.NewClient()

    fmt.Println("Books by Austen:")
    it := client.ListBooks(gutendex.Query{Author: "Austen"})
    for it.Next() {
        b := it.Value()
        fmt.Printf("%d: %s\n", b.ID, b.Title)
    }
    if err := it.Err(); err != nil {
        fmt.Println("iterator error:", err)
    }

    ctx := context.Background()
    book, err := client.GetBook(ctx, 1342)
    if err != nil {
        if gutendex.IsNotFound(err) {
            fmt.Println("book not found")
        } else {
            fmt.Println("error:", err)
        }
        return
    }
    fmt.Printf("Fetched book: %s (ID %d)\n", book.Title, book.ID)
}
```

## Filtered Search

Use the `Query` type to filter results by topic, language and other attributes.

```go
package main

import (
    "fmt"

    gutendex "github.com/alex-rs/go-gutendex"
)

func main() {
    client := gutendex.NewClient()

    it := client.ListBooks(gutendex.Query{
        Topic:    "Science Fiction",
        Language: "en",
    })
    for it.Next() {
        b := it.Value()
        fmt.Printf("%d: %s\n", b.ID, b.Title)
    }
    if err := it.Err(); err != nil {
        fmt.Println("iterator error:", err)
    }
}
```

## Fetch a Book by ID

```go
package main

import (
    "context"
    "fmt"

    gutendex "github.com/alex-rs/go-gutendex"
)

func main() {
    client := gutendex.NewClient()
    ctx := context.Background()
    book, err := client.GetBook(ctx, 1342)
    if err != nil {
        if gutendex.IsNotFound(err) {
            fmt.Println("book not found")
        } else {
            fmt.Println("error:", err)
        }
        return
    }
    fmt.Printf("Fetched book: %s (ID %d)\n", book.Title, book.ID)
}
```

## Keyword Search

The `Search` helper performs a simple author keyword search.

```go
package main

import (
    "fmt"

    gutendex "github.com/alex-rs/go-gutendex"
)

func main() {
    client := gutendex.NewClient()
    it := client.Search("Poe")
    for it.Next() {
        fmt.Println(it.Value().Title)
    }
    if err := it.Err(); err != nil {
        fmt.Println("iterator error:", err)
    }
}
```
