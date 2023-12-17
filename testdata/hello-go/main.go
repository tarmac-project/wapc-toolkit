/*
Simple example of a wapc guest module with a host callback.
*/
package main

import (
	"fmt"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	// Register the functions that can be called from the host
	// multiple functions can be registered at once.
	wapc.RegisterFunctions(wapc.Functions{
		"example": Example,
	})
}

// Example is a simple function that adheres to the wapc signature.
func Example(payload []byte) ([]byte, error) {
	// Execute Host Callback
	_, err := wapc.HostCall("namespace", "module", "function", payload)
	if err != nil {
		return []byte(""), fmt.Errorf("host callback failed -  %w", err)
	}
	return []byte("Hello World!"), nil
}
