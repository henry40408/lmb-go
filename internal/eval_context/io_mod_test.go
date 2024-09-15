package eval_context

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkRead(b *testing.B) {
	var state sync.Map
	e, _ := NewTestEvalContext(strings.NewReader(""))
	compiled, err := e.Compile(strings.NewReader(`
  local io = require('io')
  return io.read('*a')
  `), "read")
	if err != nil {
		b.Error(err)
	}
	for i := 0; i < b.N; i++ {
		e.Eval(context.Background(), compiled, &state)
	}
}

func TestRead(t *testing.T) {
	var state sync.Map
	var cases = map[string]struct {
		input    string
		format   string
		expected interface{}
	}{
		"read_1_byte":             {input: "foobar", format: "1", expected: "f"},
		"read_all_bytes":          {input: "foobar", format: "6", expected: "foobar"},
		"read_more_bytes":         {input: "foobar", format: "7", expected: "foobar"},
		"read_100_bytes":          {input: "foobar", format: "100", expected: "foobar"},
		"read_unicode_1_byte":     {input: "測試", format: "1", expected: []interface{}{230.0}},
		"read_unicode_3_bytes":    {input: "測試", format: "3", expected: "測"},
		"read_unicode_4_bytes":    {input: "測試", format: "4", expected: []interface{}{230.0, 184.0, 172.0, 232.0}},
		"read_unicode_6_bytes":    {input: "測試", format: "6", expected: "測試"},
		"read_unicode_more_bytes": {input: "測試", format: "7", expected: "測試"},
		"read_number":             {input: "1949", format: "'*n'", expected: 1949.0},
		"read_all":                {input: "hello 你好，世界 world", format: "'*a'", expected: "hello 你好，世界 world"},
		"read_line_w_eol":         {input: "line 1\nline 2", format: "'*L'", expected: "line 1\n"},
		"read_line_wo_eol":        {input: "line 1\nline 2", format: "'*l'", expected: "line 1"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e, _ := NewTestEvalContext(strings.NewReader(tc.input))
			res, err := e.EvalScript(context.Background(), fmt.Sprintf(`
      local io = require('io')
      return io.read(%s)
      `, tc.format), &state)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, res)
		})
	}
}
