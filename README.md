# wapc-toolkit

[![PkgGoDev](https://pkg.go.dev/badge/github.com/tarmac-project/wapc-toolkit)](https://pkg.go.dev/github.com/tarmac-project/wapc-toolkit)
[![Build Status](https://github.com/tarmac-project/wapc-toolkit/actions/workflows/build.yml/badge.svg)](https://github.com/tarmac-project/wapc-toolkit/actions/workflows/build.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/tarmac-project/wapc-toolkit)](https://goreportcard.com/report/github.com/tarmac-project/wapc-toolkit)
[![codecov](https://codecov.io/gh/tarmac-project/wapc-toolkit/graph/badge.svg?token=mysp7aUTWp)](https://codecov.io/gh/tarmac-project/wapc-toolkit)

The wapc-toolkit is a collection of packages for Go applications using WebAssembly Procedure Calls (waPC).

The waPC protocol standardizes communications and error handling between both the Host and Guest. It defines how WebAssembly Hosts invoke functions within the Guest and how Hosts extend host-level functionality to the Guest modules.

### What is in the toolkit?

This toolkit aims to solve common problems for Go applications using waPC. The toolkit is still growing but already has some valuable tools.

| Package | Description | Go Docs |
| --- | --- | --- |
| Callbacks | A waPC HostCall callback router, extending multiple callbacks to waPC guests. | [![PkgGoDev](https://pkg.go.dev/badge/github.com/tarmac-project/wapc-toolkit/callbacks)](https://pkg.go.dev/github.com/tarmac-project/wapc-toolkit/callbacks) |
| Engine | A simplified interface for hosts loading and executing waPC guest modules. | [![PkgGoDev](https://pkg.go.dev/badge/github.com/tarmac-project/wapc-toolkit/engine)](https://pkg.go.dev/github.com/tarmac-project/wapc-toolkit/engine) |

#### waPC Go Implementations

- Host: [wapc-go](https://github.com/wapc/wapc-go)
- Guest: [wapc-guest-tinygo](https://github.com/wapc/wapc-guest-tinygo)

### Use Cases

WebAssembly (WASM) is not just for the browser; with the creation of the WebAssembly System Interface (WASI), WebAssembly can be used as a language-agnostic way to enable many use cases. The waPC protocol and this toolkit make it approachable for everyone.

What could we use this toolkit for? Here are some thoughts.

- Stored Procedures: WASM provides a modern & testable approach to extending custom processing on databases, message brokers, or any other platform.
- Serverless runtimes: WASM solves much of the cold-start problem, and waPC enables runtimes to extend functionality to Guest modules.
- Dynamic Plugins: do you need to make your application more modular but also want to separate your core from your modules? WASM is an excellent answer to this problem.
- And many more.

## Getting Started

The toolkit is a collection of packages you can import and use independently.

The following shows an example of leveraging the engine to load and execute a waPC guest module.

```go
package main

import (
	"github.com/tarmac-project/wapc-toolkit/engine"
)

func main() {
	// Create a new engine server.
	server, err := engine.New(ServerConfig{...})
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
```

## Contributing

If you would like to contribute, please fork the repo and send in a pull request. All contributions are welcome and 
appreciated.

## License

The Apache 2 License. Please see [LICENSE](LICENSE) for more information.
