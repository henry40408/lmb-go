package executor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	m.Run()
}

func BenchmarkEval(b *testing.B) {
	var state sync.Map
	e, store := NewTestExecutor()
	compiled, _ := e.Compile(strings.NewReader("return 1"), "a")
	for i := 0; i < b.N; i++ {
		e.Eval(context.Background(), compiled, &state, store)
	}
}

func BenchmarkEvalScript(b *testing.B) {
	var state sync.Map
	e, store := NewTestExecutor()
	for i := 0; i < b.N; i++ {
		e.EvalScript(context.Background(), "return 1", &state, store)
	}
}

func TestEval(t *testing.T) {
	var state sync.Map
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
			e, store := NewTestExecutor()
			res, err := e.EvalScript(context.Background(), tc.script, &state, store)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestEvalWithTimeout(t *testing.T) {
	var state sync.Map
	e, store := NewTestExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_, err := e.EvalScript(ctx, "while true do; end", &state, store)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestEvalReader(t *testing.T) {
	var state sync.Map

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupServer(listener)
	state.Store("url", fmt.Sprintf("http://%s", listener.Addr().String()))

	matches, err := filepath.Glob("../lua-examples/*.lua")
	assert.NoError(t, err)
	for _, path := range matches {
		e, store := NewTestExecutor()
		file, err := os.Open(path)
		assert.NoError(t, err)
		defer file.Close()
		_, err = e.EvalReader(context.Background(), file, &state, store)
		assert.NoError(t, err, path)
	}
}

func TestState(t *testing.T) {
	var state sync.Map
	state.Store("a", 1.0)

	e, store := NewTestExecutor()
	res, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.state['b'] = m.state['a'] + 1
  return true
  `, &state, store)
	assert.NoError(t, err)
	assert.Equal(t, true, res)

	a, _ := state.Load("a")
	b, _ := state.Load("b")
	assert.Equal(t, 1.0, a)
	assert.Equal(t, 2.0, b)
}

func TestStore(t *testing.T) {
	var state sync.Map
	e, store := NewTestExecutor()
	res, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.store['a'] = 47
  assert(m.store['a'] == 47)
  assert(not m.store['b'])
  return true
  `, &state, store)
	assert.NoError(t, err)
	assert.Equal(t, true, res)
}

func TestStoreUpdate(t *testing.T) {
	var state sync.Map
	e, store := NewTestExecutor()

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
  `, &state, store)
	assert.Error(t, err, "insufficient fund")
	assert.Nil(t, failed)

	alice, err := store.Get("alice")
	assert.NoError(t, err)
	assert.Equal(t, float64(50), alice)
	bob, err := store.Get("bob")
	assert.NoError(t, err)
	assert.Equal(t, float64(50), bob)

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
  `, &state, store)
	assert.NoError(t, err)
	assert.Equal(t, true, success)

	alice, err = store.Get("alice")
	assert.NoError(t, err)
	assert.Equal(t, float64(0), alice)
	bob, err = store.Get("bob")
	assert.NoError(t, err)
	assert.Equal(t, float64(100), bob)
}

func setupServer(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "127.0.0.1\n")
	})
	s := &http.Server{Handler: mux}
	go s.Serve(listener)
}
