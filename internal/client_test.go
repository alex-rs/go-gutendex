package internal

import (
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
