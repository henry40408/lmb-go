package eval_context

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

func BenchmarkCompile(b *testing.B) {
	e, _ := NewTestEvalContext(strings.NewReader(""))
	for range b.N {
		e.Compile(strings.NewReader("return 1"), "a")
	}
}

func BenchmarkEvalCompiled(b *testing.B) {
	var state sync.Map
	e, _ := NewTestEvalContext(strings.NewReader(""))
	compiled, _ := e.Compile(strings.NewReader("return 1"), "a")
	for range b.N {
		_, err := e.Eval(context.Background(), compiled, &state)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEvalConcurrency(b *testing.B) {
	var state sync.Map
	e, _ := NewTestEvalContext(strings.NewReader(""))
	compiled, _ := e.Compile(strings.NewReader(`
  local m = require('lmb')
  m.store:update(function(store)
    store['counter'] = (store['counter'] or 0) + 1
  end)
  return true
  `), "concurrency")
	for range b.N {
		_, err := e.Eval(context.Background(), compiled, &state)
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkEvalScript(b *testing.B) {
	var state sync.Map
	e, _ := NewTestEvalContext(strings.NewReader(""))
	for range b.N {
		e.EvalScript(context.Background(), "return 1", &state)
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
		{"list", "return {1, 2}", []interface{}{1.0, 2.0}},
		{"table", "return {a = 1, b = 2}", map[string]interface{}{"a": 1.0, "b": 2.0}},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			e, _ := NewTestEvalContext(strings.NewReader(""))
			res, err := e.EvalScript(context.Background(), tc.script, &state)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestEvalWithTimeout(t *testing.T) {
	var state sync.Map
	e, _ := NewTestEvalContext(strings.NewReader(""))
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	_, err := e.EvalScript(ctx, "while true do; end", &state)
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
		e, _ := NewTestEvalContext(strings.NewReader(""))
		file, err := os.Open(path)
		assert.NoError(t, err)
		defer file.Close()
		_, err = e.EvalReader(context.Background(), file, &state)
		assert.NoError(t, err, path)
	}
}

func TestParse(t *testing.T) {
	e, _ := NewTestEvalContext(strings.NewReader(""))
	_, err := e.Parse(strings.NewReader("ret 1"), "invalid")
	assert.ErrorContains(t, err, "line:1(column:5)")
}

func setupServer(listener net.Listener) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "127.0.0.1\n")
	})
	s := &http.Server{Handler: mux}
	go s.Serve(listener)
}
