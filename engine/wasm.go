/*
Package engine is part of the wapc-toolkit and provides a simplified interface for loading and executing waPC guest
modules.

The wapc-toolkit is a collection of packages focused on providing tools for Go applications using the WebAssembly
Procedure Calls (waPC) standard. The waPC standard is used by host and module applications to communicate. The
Go implementation github.com/wapc/wapc-go offers multiple host runtimes supporting the WebAssembly System
Interface (WASI) and an SDK for guest modules.

This package aims to provide a simplified WebAssembly Engine Server that implements the waPC protocol. Under
the covers, this package is leveraging the github.com/wapc/wapc-go host implementation but offers an easy-to-use
interface for loading waPC guest WebAssembly modules and executing exported functions.

Use this package if you have a Go application and want to enable extended functionality via WebAssembly.

Examples of use cases could be stored procedures within a database, serverless functions, or
language-agnostic plugins.

Usage:

	import (
		"github.com/tarmac-project/wapc-toolkit/engine"
	)

	func main() {
		// Create a new engine server.
		server, err := engine.New(ServerConfig{})
		if err != nil {
			// do something
		}

		// Load the guest module.
		err = server.LoadModule(engine.ModuleConfig{
			Name:     "my-guest-module",
			Filepath: "./my-guest-module.wasm",
		})
		if err != nil {
			// do something
		}

		// Lookup the guest module.
		m, err := server.Module("my-guest-module")
		if err != nil {
			// do something
		}

		// Call the Hello function within the guest module.
		rsp, err := m.Run("Hello", []byte("world"))
		if err != nil {
			// do something
		}
	}
*/
package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	wapc "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

var (
	// ErrModuleNotFound is returned when a module is not found.
	ErrModuleNotFound = errors.New("module not found")

	// ErrCallbackNil is returned when the callback function is nil.
	ErrCallbackNil = errors.New("callback cannot be nil")
)

// ServerConfig is used to configure the initial Server.
type ServerConfig struct {

	// Callback is a user-defined function that is called when waPC guests use the HostCall function.
	// HostCalls enable waPC guests to perform a callback to the Host application. This capability
	// allows a host to expose functionality to a guest via the waPC protocol.
	//
	// The callback function is registered via the waPC runtime engine and is called with parameters
	// specified by the guest.
	Callback func(context.Context, string, string, string, []byte) ([]byte, error)
}

// Server provides the ability to load and execute waPC guest modules.
type Server struct {
	sync.RWMutex

	// callback is provided by the caller, this callback function is used when waPC guests perform a host callback.
	callback func(context.Context, string, string, string, []byte) ([]byte, error)

	// modules is a map for storing and fetching modules that have already been loaded.
	modules map[string]*Module
}

// New will create a new waPC Engine Server. The Server is a simplified interface for applications to
// load waPC guests.
//
// Once the Server is created, users can load waPC guest modules, allowing them to execute exported functions.
func New(cfg ServerConfig) (*Server, error) {
	s := &Server{}
	s.modules = make(map[string]*Module)

	if cfg.Callback == nil {
		return s, ErrCallbackNil
	}

	s.callback = cfg.Callback
	return s, nil
}

// Close will shut down the server and clean up any loaded modules, including the module pools.
func (s *Server) Close() {
	s.RLock()
	defer s.RUnlock()
	for _, m := range s.modules {
		defer m.cancel()
		defer m.module.Close(m.ctx)
		defer m.pool.Close(m.ctx)
	}
}

// LoadModule will fetch the WebAssembly Module specified by the user-provided ModuleConfig and initialize it via
// the Server.
//
// Once a Module is loaded, users can fetch the Module from the Server and call the exported functions.
func (s *Server) LoadModule(cfg ModuleConfig) error {
	if cfg.Name == "" || cfg.Filepath == "" {
		return fmt.Errorf("%w: key and file cannot be empty", ErrInvalidModuleConfig)
	}

	// Create Module
	m := &Module{
		Name: cfg.Name,
	}

	// Create context
	m.ctx, m.cancel = context.WithCancel(context.Background())

	// Set Pool Size
	m.poolSize = uint64(DefaultPoolSize)
	if cfg.PoolSize > 0 {
		m.poolSize = uint64(cfg.PoolSize)
	}

	// Read the WASM module file
	guest, err := os.ReadFile(cfg.Filepath)
	if err != nil {
		return fmt.Errorf("unable to read wasm module file - %w", err)
	}

	// Initiate waPC Engine
	engine := wazero.Engine()

	// Create a new Module from file contents
	m.module, err = engine.New(m.ctx, s.callback, guest, &wapc.ModuleConfig{
		Logger: wapc.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return fmt.Errorf("unable to load module with wasm file %s - %w", cfg.Filepath, err)
	}

	// Create pool for module
	m.pool, err = wapc.NewPool(m.ctx, m.module, m.poolSize)
	if err != nil {
		return fmt.Errorf("unable to create module pool for wasm file %s - %w", cfg.Filepath, err)
	}

	s.Lock()
	defer s.Unlock()
	s.modules[m.Name] = m

	return nil
}

// Module will return the specified Module.
//
// If the module is not found, ErrModuleNotFound will be returned.
func (s *Server) Module(key string) (*Module, error) {
	s.RLock()
	defer s.RUnlock()
	if m, ok := s.modules[key]; ok {
		return m, nil
	}
	return &Module{}, ErrModuleNotFound
}
