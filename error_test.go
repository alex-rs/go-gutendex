package gutendex

import (
	"errors"
	"testing"
)

func TestErrorErrorFormatting(t *testing.T) {
	t.Run("without inner error", func(t *testing.T) {
		e := &Error{Op: "op"}
		if got := e.Error(); got != "op" {
			t.Fatalf("unexpected: %q", got)
		}
	})
	t.Run("with inner error", func(t *testing.T) {
		inner := errors.New("boom")
		e := &Error{Op: "op", Err: inner}
		if got := e.Error(); got != "op: boom" {
			t.Fatalf("unexpected: %q", got)
		}
	})
}

func TestErrorIs(t *testing.T) {
	inner := errors.New("x")
	e := &Error{Kind: ErrNotFound, Err: inner}
	if !errors.Is(e, &Error{Kind: ErrNotFound}) {
		t.Fatalf("expected Is to match kind")
	}
	if errors.Is(e, &Error{Kind: ErrServer}) {
		t.Fatalf("unexpected match")
	}
}

func TestErrorUnwrap(t *testing.T) {
	inner := errors.New("x")
	e := &Error{Err: inner}
	if got := errors.Unwrap(e); got != inner {
		t.Fatalf("unwrap mismatch")
	}
}
