package callbacks

import (
	"context"
	"sync"
	"testing"
	"time"
)

type Counter struct {
	sync.RWMutex
	count int
}

func (c *Counter) Increment() {
	c.Lock()
	defer c.Unlock()
	c.count++
}

func (c *Counter) Value() int {
	c.RLock()
	defer c.RUnlock()
	return c.count
}

func TestRouter(t *testing.T) {
	// Track the number of times the callback is called
	callbackCounter := &Counter{}

	// Track the number of times the prefunc is called
	prefuncCounter := &Counter{}

	// Track the number of times the postfunc is called
	postfuncCounter := &Counter{}

	// Create a new router
	router := New(RouterConfig{
		PreFunc: func(_ CallbackRequest) ([]byte, error) {
			prefuncCounter.Increment()
			return nil, nil
		},
		PostFunc: func(_ CallbackResult) {
			postfuncCounter.Increment()
		},
	})

	// Define a callback
	cbCfg := CallbackConfig{
		Namespace:  "default",
		Capability: "counter",
		Operation:  "increment",
		Func: func(data []byte) ([]byte, error) {
			callbackCounter.Increment()
			return nil, nil
		},
	}

	t.Run("Register Callback", func(t *testing.T) {
		err := router.RegisterCallback(cbCfg)
		if err != nil {
			t.Fatalf("Unexpected error registering callback: %s", err)
		}

		t.Run("Lookup Callback", func(t *testing.T) {
			cb, err := router.Lookup("default", "counter", "increment")
			if err != nil {
				t.Fatalf("Unexpected error looking up callback: %s", err)
			}

			if cb.Namespace != cbCfg.Namespace {
				t.Errorf("Unexpected namespace: %s", cb.Namespace)
			}
		})
	})

	t.Run("Callback", func(t *testing.T) {
		_, err := router.Callback(context.Background(), "default", "counter", "increment", []byte(""))
		if err != nil {
			t.Errorf("Unexpected error calling callback: %s", err)
		}

		t.Run("Validate Callback Execution", func(t *testing.T) {
			if callbackCounter.Value() != 1 {
				t.Errorf("Unexpected callback count: %d, expected: 1", callbackCounter.Value())
			}
		})

		t.Run("Validate PreFunc Execution", func(t *testing.T) {
			if prefuncCounter.Value() != 1 {
				t.Errorf("Unexpected prefunc count: %d, expected: 1", prefuncCounter.Value())
			}
		})

		t.Run("Validate PostFunc Execution", func(t *testing.T) {
			if postfuncCounter.Value() != 1 {
				t.Errorf("Unexpected postfunc count: %d, expected: 1", postfuncCounter.Value())
			}
		})
	})

	t.Run("Callback with Expired Context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		cancel()
		_, err := router.Callback(ctx, "default", "counter", "increment", []byte(""))
		if err != ErrCanceled {
			t.Errorf("Expected canceled error calling callback, got: %s", err)
		}
	})

	t.Run("Unregister Callback", func(t *testing.T) {
		err := router.UnregisterCallback(cbCfg)
		if err != nil {
			t.Errorf("Unexpected error unregistering callback: %s", err)
		}

		t.Run("Lookup Unregistered Callback", func(t *testing.T) {
			_, err := router.Lookup("default", "counter", "increment")
			if err != ErrNotFound {
				t.Errorf("Expected notfound error looking up callback, got: %s", err)
			}
		})

		t.Run("Callback expecting error", func(t *testing.T) {
			_, err := router.Callback(context.Background(), "default", "counter", "increment", []byte(""))
			if err != ErrNotFound {
				t.Errorf("Expected notfound error calling callback, got: %s", err)
			}
		})
	})

	t.Run("Unregister an unregistered callback", func(t *testing.T) {
		err := router.UnregisterCallback(cbCfg)
		if err != nil {
			t.Errorf("Unexpected error unregistering callback: %s", err)
		}
	})
}
