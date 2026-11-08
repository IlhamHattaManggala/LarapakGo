package event

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestEventDispatcherSync(t *testing.T) {
	d := New()

	var order []string
	d.Listen("user.registered", func(ev string, payload any) error {
		p := payload.(map[string]any)
		order = append(order, "listener1:"+p["name"].(string))
		return nil
	})

	d.Listen("user.registered", func(ev string, payload any) error {
		order = append(order, "listener2")
		return errors.New("simulated error")
	})

	errs := d.Dispatch("user.registered", map[string]any{"name": "Slamet"})
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errs))
	}
	if len(order) != 2 || order[0] != "listener1:Slamet" || order[1] != "listener2" {
		t.Fatalf("unexpected execution order: %v", order)
	}

	// Test Forget
	d.Forget("user.registered")
	if d.HasListeners("user.registered") {
		t.Fatal("expected no listeners after forget")
	}
}

func TestEventDispatcherAsync(t *testing.T) {
	d := New()

	var wg sync.WaitGroup
	wg.Add(1)

	var received string
	d.Listen("order.created", func(ev string, payload any) error {
		defer wg.Done()
		received = payload.(string)
		return nil
	})

	d.DispatchAsync("order.created", "ORD-12345")

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if received != "ORD-12345" {
			t.Fatalf("expected ORD-12345, got %s", received)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("async listener timed out")
	}
}
