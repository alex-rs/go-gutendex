package gutendex

import (
	"errors"
	"testing"
)

func TestErrorFormatting(t *testing.T) {
	e := &Error{Op: "op"}
	if e.Error() != "op" {
		t.Fatalf("unexpected: %q", e.Error())
	}
	inner := errors.New("boom")
	e = &Error{Op: "op", Err: inner}
	if e.Error() != "op: boom" {
		t.Fatalf("unexpected: %q", e.Error())
	}
}

func TestErrorIsAndUnwrap(t *testing.T) {
	inner := errors.New("x")
	e := &Error{Kind: ErrNotFound, Err: inner}
	if !errors.Is(e, &Error{Kind: ErrNotFound}) {
		t.Fatalf("expected Is to match kind")
	}
	if errors.Is(e, &Error{Kind: ErrServer}) {
		t.Fatalf("unexpected match")
	}
	if errors.Unwrap(e) != inner {
		t.Fatalf("unwrap mismatch")
	}
}
