# Go Gutendex

[![Go Report Card](https://goreportcard.com/badge/github.com/alex-rs/go-gutendex)](https://goreportcard.com/report/github.com/alex-rs/go-gutendex)

A small Go SDK for the [Gutendex](https://gutendex.com/) API.

## Install

```
go get github.com/alex-rs/go-gutendex
```

## Quick Start

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

