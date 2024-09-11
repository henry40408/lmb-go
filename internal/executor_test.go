package internal

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
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
			e := NewExecutor()
			res, err := e.Eval(context.Background(), tc.script)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestEvalWithTimeout(t *testing.T) {
	e := NewExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_, err := e.Eval(ctx, "while true do; end")
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestEvalFile(t *testing.T) {
	matches, err := filepath.Glob("../lua-examples/*.lua")
	assert.NoError(t, err)
	for _, path := range matches {
		e := NewExecutor()
		_, err := e.EvalFile(context.Background(), path)
		assert.NoError(t, err, path)
	}
}
