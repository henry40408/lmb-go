package eval_context

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	var state sync.Map
	state.Store("a", 1.0)

	e, _ := NewTestEvalContext(strings.NewReader(""))
	res, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.state['b'] = m.state['a'] + 1
  return true
  `, &state)
	assert.NoError(t, err)
	assert.Equal(t, true, res)

	a, _ := state.Load("a")
	b, _ := state.Load("b")
	assert.Equal(t, 1.0, a)
	assert.Equal(t, 2.0, b)
}
