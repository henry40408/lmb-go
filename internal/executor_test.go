package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExectue(t *testing.T) {
	testCases := []struct {
		name     string
		script   string
		expected interface{}
	}{
		{"nil", "", nil},
		{"bool", "return true", bool(true)},
		{"number", "return 1", float64(1)},
		{"string", "return 'hello'", string("hello")},
		{"list", "return {1, 2}", []interface{}{float64(1), float64(2)}},
		{"table", "return {a = 1, b = 2}", map[string]interface{}{"a": float64(1), "b": float64(2)}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			s := Executor{Script: tc.script}
			res, err := s.Eval()
			assert.NoError(t, err)
			assert.Equal(t, res, tc.expected, res, "expected %v to be %v", res, tc.expected)
		})
	}
}
