/*
Package callbacks provides a callback router as part of the wapc-toolkit; users can use the callbacks
package with other packages in the toolkit or directly with the github.com/wapc/wapc-go package.

The wapc-toolkit is a collection of packages focused on providing tools for Go applications using
the WebAssembly Procedure Calls (waPC) standard. The waPC standard is used by host and module
applications to communicate. The Go implementation github.com/wapc/wapc-go offers multiple host
runtimes supporting the WebAssembly System Interface (WASI) and an SDK for guest modules.

The Go waPC package allows modules to perform callbacks to hosts via a hostcall function.
When a host initiates the waPC engine, it can register a single function to handle these host calls.

The callbacks package provides a router that can be registered with the waPC engine. It enables
routing host calls based on the Namespace, Capability, and Operation specified by the guest module.
Hosts can extend many different capabilities to guest modules with the callback router.
*/
package callbacks

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrCanceled is returned when the callback context is canceled or expired.
	//
	// Context is checked before calling the callback function. If the context is canceled
	// or expired, the router will return this error and not execute the callback function.
	ErrCanceled = errors.New("context canceled or expired")

	// ErrNotFound is returned when the callback function is not found.
	//
	// The router will not execute any PreFunc or PostFunc functions if the callback function
	// is not found.
	ErrNotFound = errors.New("callback not found")

	// ErrCallbackExists is returned when the callback already exists.
	ErrCallbackExists = errors.New("callback already exists")
)

// RouterConfig is a configuration struct used to create a new Router instance.
type RouterConfig struct {
	// PreFunc is a user-defined function registered to a router instance and called before
	// callback function execution.
	//
	// This function aims to enable middleware-like functionality for the callback router.
	// Users can use PreFunc for logging, metrics, or any security-based validations that
	// should be executed and checked before calling the registered callback.
	//
	// If the PreFunc function returns an error, the router will return the error and
	// response payload to the caller and abandon any attempt to call the registered
	// callback function.
	//
	// If a callback execution is for an unknown function, the router will return
	// a not found error and not execute the PreFunc function.
	PreFunc func(CallbackRequest) ([]byte, error)

	// PostFunc is a user-defined function registered to a router instance and called after
	// callback function execution.
	//
	// This function aims to enable middleware-like functionality for the callback router.
	//
	// Users can use PostFunc for logging, metrics, or any post-callback validations.
	//
	// If a callback execution is for an unknown function, the router will return a not found
	// error and not execute the PostFunc function.
	PostFunc func(CallbackResult)
}

// Router is a callback router that enables users to register callback functions and execute
// them by providing a namespace, capability, and operation.
type Router struct {
	sync.RWMutex

	// callbacks is a map of registered callbacks. The key is a string of the form
	// namespace:capability:operation.
	callbacks map[string]*Callback

	// preFunc is a user-defined function registered to a router instance and called before
	// callback function execution. See RouterConfig for more details.
	preFunc func(CallbackRequest) ([]byte, error)

	// postFunc is a user-defined function registered to a router instance and called after
	// callback function execution. See RouterConfig for more details.
	postFunc func(CallbackResult)
}

// New creates a new Router instance.
func New(cfg RouterConfig) (*Router, error) {
	r := &Router{
		callbacks: make(map[string]*Callback),
		preFunc:   cfg.PreFunc,
		postFunc:  cfg.PostFunc,
	}
	return r, nil
}

// Close clears the router's callback map and shuts down the router.
func (r *Router) Close() {
	// Lock router
	r.Lock()
	defer r.Unlock()

	// Clear callbacks map
	r.callbacks = make(map[string]*Callback)
}

// RegisterCallback adds a callback to the router. If the callback already exists, an error
// is returned.
func (r *Router) RegisterCallback(cfg CallbackConfig) error {
	// Validate Config
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Check if callback already exists
	if _, err := r.Lookup(cfg.Namespace, cfg.Capability, cfg.Operation); err == nil {
		return ErrCallbackExists
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

// UnregisterCallback removes a callback from the router. If the callback does not exist,
// no error is returned.
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

// Callback executes callbacks registered to the router. It will identify the Callback by
// the user-provided Namespace, Capability, and Operation and execute the associated function,
// passing the provided input to the callback function.
//
// If any PreFunc functions are defined, Callback will execute them before executing the identified Callback.
//
// After execution, the router will call any PostFunc functions defined.
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

// Lookup returns a copy of the callback function registered to the router.
// If the callback function is not found, the function returns ErrNotFound.
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
