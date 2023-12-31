package callbacks

import (
	"errors"
	"time"
)

var (
	// ErrInvalidNamespace is returned when the callback namespace is invalid.
	ErrInvalidNamespace = errors.New("invalid namespace")

	// ErrInvalidCapability is returned when the callback capability is invalid.
	ErrInvalidCapability = errors.New("invalid capability")

	// ErrInvalidOperation is returned when the callback operation is invalid.
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrInvalidFunc is returned when the callback function is invalid.
	ErrInvalidFunc = errors.New("invalid func: cannot be nil")
)

// CallbackConfig is the user-provided configuration for a callback.
// It is used to register a callback with the callback router.
//
// Namespace, Capability, and Operation are used to identify a callback.
type CallbackConfig struct {
	// Namespace represents the namespace a callback is registered to.
	Namespace string

	// Capability represents the capability a callback is registered to.
	Capability string

	// Operation represents the operation a callback performs.
	Operation string

	// Func is the callback function that will be called when a callback is triggered.
	Func func(input []byte) ([]byte, error)
}

// Validate validates the callback configuration. It returns an error if the configuration
// values are invalid or missing any required fields.
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

// Callback represents a callback registered with the callback router.
type Callback struct {
	// Namespace represents the namespace a callback is registered to.
	Namespace string

	// Capability represents the capability a callback is registered to.
	Capability string

	// Operation represents the operation a callback performs.
	Operation string

	// Func is the callback function that will be called when a callback is triggered.
	Func func(input []byte) ([]byte, error)
}

// CallbackRequest represents a callback request made to the callback router.
type CallbackRequest struct {
	// Namespace is the user-provided namespace for the callback request.
	Namespace string

	// Capability is the user-provided capability for the callback request.
	Capability string

	// Operation is the user-provided operation for the callback request.
	Operation string

	// Input is the user-provided input for the callback request.
	Input []byte

	// StartTime is the time the callback router receives the callback request.
	// The callback router sets this time before calling any pre-function hooks.
	// This time may differ from when the WASM module made the callback request.
	StartTime time.Time
}

// CallbackResult represents the result of a callback request. It is provided to
// any PostFunc hooks registered with the callback router.
type CallbackResult struct {
	// Namespace is the user-provided namespace for the callback request.
	Namespace string

	// Capability is the user-provided capability for the callback request.
	Capability string

	// Operation is the user-provided operation for the callback request.
	Operation string

	// Input is the user-provided input for the callback request.
	Input []byte

	// Output is the callback function output provided to the WASM module.
	Output []byte

	// Err is the error returned by the callback function provided to the WASM module.
	Err error

	// StartTime is the time the callback router receives the callback request.
	// The callback router sets this time before calling any pre-function hooks.
	// This time may differ from when the WASM module made the callback request.
	StartTime time.Time

	// EndTime is the time the callback router finishes processing the callback request.
	// The callback router sets this time before calling any post-function hooks and returning
	// the response to the WASM module.
	EndTime time.Time
}
