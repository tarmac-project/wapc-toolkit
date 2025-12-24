package engine

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tetratelabs/wazero/sys"
)

func TestIgnoreModuleExitCodeZero(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		inputErr error
		expect   bool
	}{
		{
			name:     "wrapped exit zero",
			inputErr: fmt.Errorf("error invoking guest: %w", sys.NewExitError(0)),
			expect:   true,
		},
		{
			name:     "bare exit zero",
			inputErr: sys.NewExitError(0),
			expect:   true,
		},
		{
			name:     "non-zero exit",
			inputErr: fmt.Errorf("call failed: %w", sys.NewExitError(5)),
			expect:   false,
		},
		{
			name:     "non exit error",
			inputErr: errors.New("boom"),
			expect:   false,
		},
	}

	for _, tc := range testCases {
		c := tc
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			if got := ignoreModuleExitCodeZero(c.inputErr); got != c.expect {
				t.Fatalf("unexpected result: got %t want %t", got, c.expect)
			}
		})
	}
}
