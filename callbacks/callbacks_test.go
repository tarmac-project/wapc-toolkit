package callbacks

import (
	"bytes"
	"context"
	"fmt"
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

var ErrTestError = fmt.Errorf("test error")

type RouterTestCase struct {
	Name              string
	RouterCfg         RouterConfig
	RouterErr         error
	EmptyPreFunc      bool
	EmptyPostFunc     bool
	ErrPreFunc        error
	CallbackCfg       CallbackConfig
	CallbackRegErr    error
	CallbackErr       error
	EmptyCallbackFunc bool
	ErrCallbackFunc   error
	CallbackInput     []byte
}

func TestRouter(t *testing.T) {
	tt := []RouterTestCase{
		// Happy path
		{
			Name:      "Happy path",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackInput: []byte("Hello World"),
		},
		// Register empty router
		{
			Name:          "Register empty router",
			RouterCfg:     RouterConfig{},
			EmptyPreFunc:  true,
			EmptyPostFunc: true,
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackInput: []byte("Hello World"),
		},
		// Register router no prefunc
		{
			Name:         "Register router no prefunc",
			RouterCfg:    RouterConfig{},
			EmptyPreFunc: true,
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackInput: []byte("Hello World"),
		},
		// Register router no postfunc
		{
			Name:          "Register router no postfunc",
			RouterCfg:     RouterConfig{},
			EmptyPostFunc: true,
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackInput: []byte("Hello World"),
		},
		// Register router with prefunc that errors
		{
			Name:       "Register router with prefunc that errors",
			RouterCfg:  RouterConfig{},
			ErrPreFunc: ErrTestError,
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackInput: []byte("Hello World"),
			CallbackErr:   ErrTestError,
		},
		// Register empty callback
		{
			Name:           "Register empty callback",
			RouterCfg:      RouterConfig{},
			CallbackCfg:    CallbackConfig{},
			CallbackRegErr: ErrInvalidNamespace,
		},
		// Register callback with empty namespace
		{
			Name:      "Register callback with empty namespace",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackRegErr: ErrInvalidNamespace,
		},
		// Register callback with empty capability
		{
			Name:      "Register callback with empty capability",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace: "default",
				Operation: "increment",
			},
			CallbackRegErr: ErrInvalidCapability,
		},
		// Register callback with empty operation
		{
			Name:      "Register callback with empty operation",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
			},
			CallbackRegErr: ErrInvalidOperation,
		},
		// Register callback with empty function
		{
			Name:      "Register callback with empty function",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			EmptyCallbackFunc: true,
			CallbackRegErr:    ErrInvalidFunc,
		},
		// Register callback that errors
		{
			Name:      "Register callback that errors",
			RouterCfg: RouterConfig{},
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
			},
			CallbackErr:     ErrTestError,
			ErrCallbackFunc: ErrTestError,
			CallbackInput:   []byte("Hello World"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			// Create counters
			callbackCounter := &Counter{}
			prefuncCounter := &Counter{}
			postfuncCounter := &Counter{}

			// Create router config
			routerCfg := tc.RouterCfg
			if !tc.EmptyPreFunc {
				routerCfg.PreFunc = func(rq CallbackRequest) ([]byte, error) {
					// Validate input
					if bytes.Compare(rq.Input, tc.CallbackInput) != 0 {
						t.Errorf("Unexpected preFunc input: %s, expected: %s", rq.Input, tc.CallbackInput)
					}
					prefuncCounter.Increment()
					return nil, tc.ErrPreFunc
				}
			}
			if !tc.EmptyPostFunc {
				routerCfg.PostFunc = func(res CallbackResult) {
					// Validate input
					if bytes.Compare(res.Input, tc.CallbackInput) != 0 {
						t.Errorf("Unexpected postFunc input: %s, expected: %s", res.Input, tc.CallbackInput)
					}
					postfuncCounter.Increment()
				}
			}

			// Create a new router
			router, err := New(routerCfg)
			if err != nil {
				if err != tc.RouterErr {
					t.Fatalf("Unexpected error creating router: %s", err)
				}
				return
			}
			if err != tc.RouterErr {
				t.Fatal("Expected error creating router")
			}

			t.Run("Lookup non-existent callback", func(t *testing.T) {
				_, err := router.Lookup(tc.CallbackCfg.Namespace, tc.CallbackCfg.Capability, tc.CallbackCfg.Operation)
				if err != ErrNotFound {
					t.Fatalf("Unexpected error looking up callback: %s", err)
				}
			})

			if tc.CallbackRegErr != nil {
				t.Run("Unregister callback with invalid config", func(t *testing.T) {
					err := router.UnregisterCallback(tc.CallbackCfg)
					if err != tc.CallbackRegErr {
						t.Fatalf("Unexpected error unregistering callback: %s", err)
					}
				})
			}

			// Define a callback
			cbCfg := tc.CallbackCfg
			if !tc.EmptyCallbackFunc {
				cbCfg.Func = func(input []byte) ([]byte, error) {
					// Validate input
					if bytes.Compare(input, tc.CallbackInput) != 0 {
						t.Errorf("Unexpected callback input: %s, expected: %s", input, tc.CallbackInput)
					}
					callbackCounter.Increment()
					return input, tc.ErrCallbackFunc
				}
			}

			t.Run("Register Callback", func(t *testing.T) {
				err := router.RegisterCallback(cbCfg)
				if err != nil {
					if err != tc.CallbackRegErr {
						t.Fatalf("Unexpected error registering callback: %s", err)
					}
					return
				}
				if err != tc.CallbackRegErr {
					t.Fatal("Expected error registering callback")
				}

				t.Run("Lookup Callback", func(t *testing.T) {
					cb, err := router.Lookup(cbCfg.Namespace, cbCfg.Capability, cbCfg.Operation)
					if err != nil {
						t.Fatalf("Unexpected error looking up callback: %s", err)
					}

					if cb.Namespace != cbCfg.Namespace {
						t.Errorf("Unexpected namespace: %s", cb.Namespace)
					}
				})

				t.Run("Try to Register Callback Again", func(t *testing.T) {
					err := router.RegisterCallback(cbCfg)
					if err != nil {
						if err != ErrCallbackExists {
							t.Fatalf("Unexpected error registering callback: %s", err)
						}
						return
					}
					t.Fatal("Expected error registering callback")
				})

				t.Run("Callback", func(t *testing.T) {
					rsp, err := router.Callback(context.Background(),
						tc.CallbackCfg.Namespace,
						tc.CallbackCfg.Capability,
						tc.CallbackCfg.Operation,
						tc.CallbackInput)
					if err != nil {
						if err != tc.CallbackErr {
							t.Fatalf("Unexpected error calling callback: %s", err)
						}
						return
					}
					if err != tc.CallbackErr {
						t.Fatal("Expected error calling callback")
					}

					if bytes.Compare(rsp, tc.CallbackInput) != 0 {
						t.Errorf("Unexpected callback response: %s, expected: %s", rsp, tc.CallbackInput)
					}

					t.Run("Validate Callback Execution", func(t *testing.T) {
						if callbackCounter.Value() != 1 {
							t.Errorf("Unexpected callback count: %d, expected: 1", callbackCounter.Value())
						}
					})

					t.Run("Validate PreFunc Execution", func(t *testing.T) {
						if !tc.EmptyPreFunc && prefuncCounter.Value() != 1 {
							t.Errorf("Unexpected prefunc count: %d, expected: 1", prefuncCounter.Value())
						}
					})

					t.Run("Validate PostFunc Execution", func(t *testing.T) {
						// Wait for goroutine to finish
						<-time.After(100 * time.Millisecond)
						if !tc.EmptyPostFunc && postfuncCounter.Value() != 1 {
							t.Errorf("Unexpected postfunc count: %d, expected: 1", postfuncCounter.Value())
						}
					})
				})

				t.Run("Callback with Expired Context", func(t *testing.T) {
					ctx, cancel := context.WithCancel(context.Background())
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

					t.Run("Unregister an unregistered callback", func(t *testing.T) {
						err := router.UnregisterCallback(cbCfg)
						if err != nil {
							t.Errorf("Unexpected error unregistering callback: %s", err)
						}
					})
				})
			})
		})
	}
}

type CallbackConfigTestCase struct {
	Name        string
	CallbackCfg CallbackConfig
	Err         error
}

func TestCallbackConfigValidation(t *testing.T) {
	tt := []CallbackConfigTestCase{
		{
			Name: "Valid CallbackConfig",
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
				Func: func(input []byte) ([]byte, error) {
					return input, nil
				},
			},
			Err: nil,
		},
		{
			Name: "Invalid Namespace",
			CallbackCfg: CallbackConfig{
				Namespace:  "",
				Capability: "counter",
				Operation:  "increment",
				Func: func(input []byte) ([]byte, error) {
					return input, nil
				},
			},
			Err: ErrInvalidNamespace,
		},
		{
			Name: "Invalid Capability",
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "",
				Operation:  "increment",
				Func: func(input []byte) ([]byte, error) {
					return input, nil
				},
			},
			Err: ErrInvalidCapability,
		},
		{
			Name: "Invalid Operation",
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "",
				Func: func(input []byte) ([]byte, error) {
					return input, nil
				},
			},
			Err: ErrInvalidOperation,
		},
		{
			Name: "Invalid Func",
			CallbackCfg: CallbackConfig{
				Namespace:  "default",
				Capability: "counter",
				Operation:  "increment",
				Func:       nil,
			},
			Err: ErrInvalidFunc,
		},
	}

	for _, tc := range tt {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.CallbackCfg.Validate()
			if err != tc.Err {
				t.Errorf("Unexpected error validating callback config: %s", err)
			}
		})
	}
}
