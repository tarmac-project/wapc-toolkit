package callbacks

import (
	"time"
)

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
