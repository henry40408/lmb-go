package executor

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/henry40408/lmb/internal/database"
	"github.com/stretchr/testify/assert"
)

func BenchmarkEval(b *testing.B) {
	var state sync.Map
	e := NewTestExecutor()
	db, err := database.OpenDB(":memory:")
	if err != nil {
		log.Fatal(err)
	}
	compiled, _ := e.Compile(strings.NewReader("return 1"), "a")
	for i := 0; i < b.N; i++ {
		e.Eval(context.Background(), compiled, &state, db)
	}
}

func BenchmarkEvalScript(b *testing.B) {
	var state sync.Map
	e := NewTestExecutor()
	db, err := database.OpenDB(":memory:")
	if err != err {
		log.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		e.EvalScript(context.Background(), "return 1", &state, db)
	}
}

func TestEval(t *testing.T) {
	var state sync.Map
	db, err := database.OpenDB(":memory:")
	assert.NoError(t, err)
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
			e := NewTestExecutor()
			res, err := e.EvalScript(context.Background(), tc.script, &state, db)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestEvalWithTimeout(t *testing.T) {
	var state sync.Map
	db, err := database.OpenDB(":memory:")
	assert.NoError(t, err)
	e := NewTestExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_, err = e.EvalScript(ctx, "while true do; end", &state, db)
	assert.Contains(t, err.Error(), "context deadline exceeded")
}

func TestEvalFile(t *testing.T) {
	var state sync.Map

	db, err := database.OpenDB(":memory:")
	assert.NoError(t, err)

	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	setupServer(listener)
	state.Store("url", fmt.Sprintf("http://%s", listener.Addr().String()))

	matches, err := filepath.Glob("../lua-examples/*.lua")
	assert.NoError(t, err)
	for _, path := range matches {
		e := NewTestExecutor()
		_, err := e.EvalFile(context.Background(), path, &state, db)
		assert.NoError(t, err, path)
	}
}

func TestState(t *testing.T) {
	var state sync.Map
	state.Store("a", 1.0)

	db, err := database.OpenDB(":memory:")
	assert.NoError(t, err)

	e := NewTestExecutor()
	res, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.state['b'] = m.state['a'] + 1
  return true
  `, &state, db)
	assert.NoError(t, err)
	assert.Equal(t, true, res)

	a, _ := state.Load("a")
	b, _ := state.Load("b")
	assert.Equal(t, 1.0, a)
	assert.Equal(t, 2.0, b)
}

func TestStore(t *testing.T) {
	var state sync.Map
	db, err := database.OpenDB(":memory:")
	assert.NoError(t, err)
	e := NewTestExecutor()
	res, err := e.EvalScript(context.Background(), `
  local m = require('lmb')
  m.store['a'] = 47
  assert(m.store['a'] == 47)
  assert(not m.store['b'])
  return true
  `, &state, db)
	assert.NoError(t, err)
	assert.Equal(t, true, res)
}

func setupServer(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "127.0.0.1\n")
	})
	s := &http.Server{Handler: mux}
	go s.Serve(listener)
}
