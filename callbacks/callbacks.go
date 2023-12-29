package callbacks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrCanceled          = errors.New("context canceled or expired")
	ErrNotFound          = errors.New("callback not found")
	ErrInvalidNamespace  = errors.New("invalid namespace")
	ErrInvalidCapability = errors.New("invalid capability")
	ErrInvalidOperation  = errors.New("invalid operation")
	ErrInvalidFunc       = errors.New("invalid func cannot be nil")
)

type RouterConfig struct {
	PreFunc  func(CallbackRequest) ([]byte, error)
	PostFunc func(CallbackResult)
}

type Router struct {
	sync.RWMutex

	callbacks map[string]*Callback
	preFunc   func(CallbackRequest) ([]byte, error)
	postFunc  func(CallbackResult)
}

func New(cfg RouterConfig) (*Router, error) {
	r := &Router{
		callbacks: make(map[string]*Callback),
		preFunc:   cfg.PreFunc,
		postFunc:  cfg.PostFunc,
	}
	return r, nil
}

func (r *Router) RegisterCallback(cfg CallbackConfig) error {
	// Validate Config
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Lock router
	r.Lock()
	defer r.Unlock()

	// Add callback to map
	r.callbacks[fmt.Sprintf("%s:%s:%s", cfg.Namespace, cfg.Capability, cfg.Operation)] = &Callback{
		Namespace:  cfg.Namespace,
		Capability: cfg.Capability,
		Operation:  cfg.Operation,
		Func:       cfg.Func,
	}

	return nil
}

func (r *Router) UnregisterCallback(cfg CallbackConfig) error {
	// Validate Config
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Lock router
	r.Lock()
	defer r.Unlock()

	// Remove callback from map
	delete(r.callbacks, fmt.Sprintf("%s:%s:%s", cfg.Namespace, cfg.Capability, cfg.Operation))

	return nil
}

func (r *Router) Callback(ctx context.Context, namespace, capability, operation string, input []byte) ([]byte, error) {
	// Validate Context
	if ctx.Err() != nil {
		return nil, ErrCanceled
	}

	// Create callback request
	req := CallbackRequest{
		Namespace:  namespace,
		Capability: capability,
		Operation:  operation,
		Input:      input,
		StartTime:  time.Now(),
	}

	// Create lookup key
	key := fmt.Sprintf("%s:%s:%s", namespace, capability, operation)

	// Read lock router
	r.RLock()
	defer r.RUnlock()

	// Lookup callback
	if cb, ok := r.callbacks[key]; ok {
		// Call preFunc
		if r.preFunc != nil {
			rsp, err := r.preFunc(req)
			if err != nil {
				// return error to caller
				return rsp, err
			}
		}

		// Call callback func
		cbRsp, err := cb.Func(input)

		// Call postFunc
		if r.postFunc != nil {
			go r.postFunc(CallbackResult{
				Namespace:  namespace,
				Capability: capability,
				Operation:  operation,
				Input:      input,
				Output:     cbRsp,
				Err:        err,
				StartTime:  req.StartTime,
				EndTime:    time.Now(),
			})
		}

		// Return output and error
		return cbRsp, err
	}

	// Return not found error
	return nil, ErrNotFound
}

func (r *Router) Lookup(namespace, capability, operation string) (Callback, error) {
	// Create lookup key
	key := fmt.Sprintf("%s:%s:%s", namespace, capability, operation)

	// Read lock router
	r.RLock()
	defer r.RUnlock()

	// Lookup callback
	if cb, ok := r.callbacks[key]; ok {
		// Create copy of callback
		cp := Callback{
			Namespace:  cb.Namespace,
			Capability: cb.Capability,
			Operation:  cb.Operation,
			Func:       cb.Func,
		}
		return cp, nil
	}

	// Return not found error
	return Callback{}, ErrNotFound
}

type CallbackConfig struct {
	Namespace  string
	Capability string
	Operation  string
	Func       func(input []byte) ([]byte, error)
}

func (c CallbackConfig) Validate() error {
	// Verify Namespace
	if c.Namespace == "" {
		return ErrInvalidNamespace
	}

	// Verify Capability
	if c.Capability == "" {
		return ErrInvalidCapability
	}

	// Verify Operation
	if c.Operation == "" {
		return ErrInvalidOperation
	}

	// Verify Func
	if c.Func == nil {
		return ErrInvalidFunc
	}

	return nil
}

type Callback struct {
	Namespace  string
	Capability string
	Operation  string
	Func       func(input []byte) ([]byte, error)
}

type CallbackRequest struct {
	Namespace  string
	Capability string
	Operation  string
	Input      []byte
	StartTime  time.Time
}

type CallbackResult struct {
	Namespace  string
	Capability string
	Operation  string
	Input      []byte
	Output     []byte
	Err        error
	StartTime  time.Time
	EndTime    time.Time
}
