package eval_context

import (
	"context"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	var state sync.Map
	e, _ := NewTestEvalContext(strings.NewReader(""))
	res, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.store['a'] = 47
  assert(m.store['a'] == 47)
  assert(not m.store['b'])
  return true
  `, &state)
	assert.NoError(t, err)
	assert.Equal(t, true, res)
}

func TestStoreUpdate(t *testing.T) {
	var state sync.Map
	e, store := NewTestEvalContext(strings.NewReader(""))

	failed, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.store['alice'] = 50
  m.store['bob'] = 50
  m.store:update(function(store)
    local alice = store['alice']
    if alice < 100 then
      error('insufficient fund')
    end
    store['alice'] = store['alice'] - 100
    store['bob'] = store['bob'] + 100
  end)
  return true
  `, &state)
	assert.Error(t, err, "insufficient fund")
	assert.Nil(t, failed)

	alice, err := store.Get("alice")
	assert.NoError(t, err)
	assert.Equal(t, 50.0, alice)
	bob, err := store.Get("bob")
	assert.NoError(t, err)
	assert.Equal(t, 50.0, bob)

	success, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.store['alice'] = 100
  m.store['bob'] = 0
  m.store:update(function(store)
    local alice = store['alice']
    if alice < 100 then
      error('insufficient fund')
    end
    store['alice'] = store['alice'] - 100
    store['bob'] = store['bob'] + 100
  end)
  return true
  `, &state)
	assert.NoError(t, err)
	assert.Equal(t, true, success)

	alice, err = store.Get("alice")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, alice)
	bob, err = store.Get("bob")
	assert.NoError(t, err)
	assert.Equal(t, 100.0, bob)
}

func TestStoreUpdateConcurrency(t *testing.T) {
	var state sync.Map
	e, store := NewTestEvalContext(strings.NewReader(""))

	reader := strings.NewReader(`
  local m = require('lmb')
  m.store:update(function(store)
    store['counter'] = (store['counter'] or 0) + 1
  end)
  return true
  `)
	compiled, err := e.Compile(reader, "concurrency")
	assert.NoError(t, err)

	var wg sync.WaitGroup

	count := 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			res, err := e.Eval(context.Background(), compiled, &state)
			assert.NoError(t, err)
			assert.Equal(t, true, res)
		}(i)
	}

	wg.Wait()

	value, err := store.Get("counter")
	assert.NoError(t, err)
	assert.Equal(t, float64(count), value)
}
