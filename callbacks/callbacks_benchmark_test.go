package callbacks

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

type BencharkRouterCase struct {
	Name         string
	RouterCfg    RouterConfig
	CallbackCfg  CallbackConfig
	NumCallbacks int
}

func BenchmarkRouter(b *testing.B) {
	tt := []BencharkRouterCase{
		{
			Name:      "1 Callback",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "benchmarks",
				Capability: "testing",
			},
			NumCallbacks: 1,
		},
		{
			Name:      "100 Callbacks",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "benchmarks",
				Capability: "testing",
			},
			NumCallbacks: 100,
		},
		{
			Name:      "1000 Callbacks",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "benchmarks",
				Capability: "testing",
			},
			NumCallbacks: 1000,
		},
		{
			Name:      "10000 Callbacks",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "benchmarks",
				Capability: "testing",
			},
			NumCallbacks: 10000,
		},
	}

	for _, tc := range tt {
		b.Run(tc.Name, func(b *testing.B) {
			// Create counters
			callbackCounter := &Counter{}
			preFuncCounter := &Counter{}
			postFuncCounter := &Counter{}

			// Define Functions
			tc.CallbackCfg.Func = func(_ []byte) ([]byte, error) {
				callbackCounter.Increment()
				return []byte{}, nil
			}
			tc.RouterCfg.PreFunc = func(_ CallbackRequest) ([]byte, error) {
				preFuncCounter.Increment()
				return []byte{}, nil
			}
			tc.RouterCfg.PostFunc = func(_ CallbackResult) {
				postFuncCounter.Increment()
			}

			// Create router
			router, err := New(tc.RouterCfg)
			if err != nil {
				b.Fatalf("Failed to create router: %s", err)
			}
			defer router.Close()

			// Register callbacks
			for i := 0; i < tc.NumCallbacks; i++ {
				cfg := tc.CallbackCfg
				cfg.Operation = fmt.Sprintf("operation-%d", i)
				if err := router.RegisterCallback(cfg); err != nil {
					b.Fatalf("Failed to register callback: %s", err)
				}
			}

			// Run benchmark
			b.Run("Random Callbacks", func(b *testing.B) {
				defer callbackCounter.Reset()
				defer preFuncCounter.Reset()
				defer postFuncCounter.Reset()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := router.Callback(context.Background(),
						tc.CallbackCfg.Namespace,
						tc.CallbackCfg.Capability,
						fmt.Sprintf("operation-%d", rand.Intn(tc.NumCallbacks)),
						[]byte{})
					if err != nil {
						b.Fatalf("Failed to invoke callback: %s", err)
					}
				}
				b.StopTimer()

				// Wait for postFunc goroutines to finish
				<-time.After(1 * time.Second)

				// Validate counters
				if callbackCounter.Value() != b.N {
					b.Fatalf("Callback counter mismatch: expected %d, got %d", b.N, callbackCounter.Value())
				}
				if preFuncCounter.Value() != b.N {
					b.Fatalf("PreFunc counter mismatch: expected %d, got %d", b.N, preFuncCounter.Value())
				}
				if postFuncCounter.Value() != b.N {
					b.Fatalf("PostFunc counter mismatch: expected %d, got %d", b.N, postFuncCounter.Value())
				}
			})

			// wait for counter reset
			<-time.After(1 * time.Second)

			b.Run("Single Callback", func(b *testing.B) {
				defer callbackCounter.Reset()
				defer preFuncCounter.Reset()
				defer postFuncCounter.Reset()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := router.Callback(context.Background(),
						tc.CallbackCfg.Namespace,
						tc.CallbackCfg.Capability,
						"operation-0",
						[]byte{})
					if err != nil {
						b.Fatalf("Failed to invoke callback: %s", err)
					}
				}
				b.StopTimer()

				// Wait for postFunc goroutines to finish
				<-time.After(1 * time.Second)

				// Validate counters
				if callbackCounter.Value() != b.N {
					b.Fatalf("Callback counter mismatch: expected %d, got %d", b.N, callbackCounter.Value())
				}
				if preFuncCounter.Value() != b.N {
					b.Fatalf("PreFunc counter mismatch: expected %d, got %d", b.N, preFuncCounter.Value())
				}
				if postFuncCounter.Value() != b.N {
					b.Fatalf("PostFunc counter mismatch: expected %d, got %d", b.N, postFuncCounter.Value())
				}
			})
		})
	}
}
