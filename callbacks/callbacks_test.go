package callbacks

import (
	"errors"
	"testing"
)

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
			if !errors.Is(err, tc.Err) {
				t.Errorf("Unexpected error validating callback config: %s", err)
			}
		})
	}
}
