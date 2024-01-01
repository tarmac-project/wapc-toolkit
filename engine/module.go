package engine

import (
	"context"
	"fmt"
	"time"

	wapc "github.com/wapc/wapc-go"
)

var (
	// ErrInvalidModuleConfig is returned when a ModuleConfig is invalid.
	ErrInvalidModuleConfig = fmt.Errorf("invalid module config")
)

const (
	// Default WebAssembly Module Pool Size.
	DefaultPoolSize = 100

	// Default WebAssembly Module Pool Timeout.
	DefaultPoolTimeout = 5
)

// ModuleConfig is used to configure WebAssembly Modules for the Server to load and ready for execution.
//
// A pool is created for WebAssembly Modules. As a module is called, the Server will take an available
// module from the pool and use it to execute the specified function.
type ModuleConfig struct {
	// Name is the name of the WebAssembly Module, which is used as a lookup key by the server when
	// fetching modules.
	Name string

	// Filepath is the path to load the .wasm module file from the file system.
	Filepath string

	// PoolSize is used to control the size of the WebAssembly Modules pool. Each module has its
	// own pool; for each invocation of the Run function, the module is taken from the pool and
	// re-added upon completion. The pool size should be large enough to support concurrent executions of
	// module functions.
	//
	// If PoolSize is not provided, DefaultPoolSize will be used.
	PoolSize int
}

// Module is a specific WebAssembly Module loaded via the WebAssembly Engine Server. Each WebAssembly
// module exposes unique functions that are callable via the Run method.
//
// To enable concurrency, a pool of Modules is created; see the ModuleConfig for details on
// tuning the pool.
type Module struct {
	//  Name is the name of the WebAssembly Module, which is used as a lookup key by the server when
	// fetching modules.
	Name string

	// ctx is a context used to clean up module instances.
	ctx context.Context

	// cancel is a context cancellation function used to clean up module instances.
	cancel context.CancelFunc

	// module is the loaded module, this is referenced for clean up and closure purposes.
	module wapc.Module

	// pool is the module pool created as part of loading a module. This pool is used to store and fetch
	// module instances as needed.
	pool *wapc.Pool

	// poolSize will determine the size of a module pool.
	poolSize uint64
}

// Run will fetch a WASM module from the available pool and call the user-provided function with the
// user-provided payload.
//
// Upon completion, Run will add the module back to the available pool.
func (m *Module) Run(function string, payload []byte) ([]byte, error) {
	var r []byte
	// Get a module instance from the pool
	i, err := m.pool.Get(DefaultPoolTimeout * time.Second)
	if err != nil {
		return r, fmt.Errorf("could not fetch module from pool - %w", err)
	}

	// Return the module to the pool
	defer func() {
		err := m.pool.Return(i) //nolint:govet // Ignore govet warning about shadowing err as it is not shadowed.
		if err != nil {
			defer i.Close(m.ctx)
		}
	}()

	// Invoke the module with the user-provided function and payload
	r, err = i.Invoke(m.ctx, function, payload)
	if err != nil {
		return r, err
	}

	return r, nil
}
