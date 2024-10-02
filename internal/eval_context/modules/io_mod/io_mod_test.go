package io_mod

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/henry40408/lmb/internal/eval_context/modules/testutil"
	"github.com/henry40408/lmb/internal/lua_convert"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

func BenchmarkRead(b *testing.B) {
	L := testutil.NewLuaTestState()
	defer L.Close()

	br := bufio.NewReader(strings.NewReader(""))
	var w bytes.Buffer
	L.PreloadModule("io", NewIoMod(br, &w).Loader)

	reader := strings.NewReader(`
  local io = require('io')
  return io.read('*a')
  `)
	chunk, err := parse.Parse(reader, "compiled")
	if err != nil {
		b.Error(err)
	}
	proto, err := lua.Compile(chunk, "compiled")
	if err != nil {
		b.Error(err)
	}
	for range b.N {
		L.Push(L.NewFunctionFromProto(proto))
		err := L.PCall(0, lua.MultRet, nil)
		if err != nil {
			b.Error(err)
		}
		if L.GetTop() > 0 {
			L.Pop(1) // pop the result or registry overflows
		}
	}
}

func TestRead(t *testing.T) {
	var cases = map[string]struct {
		input    string
		format   string
		expected interface{}
	}{
		"read_1_byte":             {input: "foobar", format: "1", expected: "f"},
		"read_all_bytes":          {input: "foobar", format: "6", expected: "foobar"},
		"read_more_bytes":         {input: "foobar", format: "7", expected: "foobar"},
		"read_100_bytes":          {input: "foobar", format: "100", expected: "foobar"},
		"read_unicode_1_byte":     {input: "測試", format: "1", expected: []interface{}{int64(230)}},
		"read_unicode_3_bytes":    {input: "測試", format: "3", expected: "測"},
		"read_unicode_4_bytes":    {input: "測試", format: "4", expected: []interface{}{int64(230), int64(184), int64(172), int64(232)}},
		"read_unicode_6_bytes":    {input: "測試", format: "6", expected: "測試"},
		"read_unicode_more_bytes": {input: "測試", format: "7", expected: "測試"},
		"read_number":             {input: "1949", format: "'*n'", expected: int64(1949)},
		"read_all":                {input: "hello 你好，世界 world", format: "'*a'", expected: "hello 你好，世界 world"},
		"read_line_w_eol":         {input: "line 1\nline 2", format: "'*L'", expected: "line 1\n"},
		"read_line_wo_eol":        {input: "line 1\nline 2", format: "'*l'", expected: "line 1"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			L := testutil.NewLuaTestState()
			defer L.Close()

			br := bufio.NewReader(strings.NewReader(tc.input))
			var w bytes.Buffer
			L.PreloadModule("io", NewIoMod(br, &w).Loader)

			script := fmt.Sprintf(`
      local io = require('io')
      return io.read(%s)
      `, tc.format)

			err := L.DoString(script)
			assert.NoError(t, err)
			assert.Greater(t, L.GetTop(), 0, "expect result")

			res := lua_convert.FromLuaValue(L.Get(-1))
			assert.Equal(t, tc.expected, res)
		})
	}
}

func TestWrite(t *testing.T) {
	var cases = map[string]struct {
		output   string
		expected []byte
	}{
		"write_char":    {output: "'a'", expected: []byte("a")},
		"write_unicode": {output: "'世界'", expected: []byte("世界")},
		"write_int":     {output: "1", expected: []byte("1")},
		"write_float":   {output: "1.23", expected: []byte("1.23")},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			L := testutil.NewLuaTestState()
			defer L.Close()

			br := bufio.NewReader(strings.NewReader(""))
			var w bytes.Buffer
			L.PreloadModule("io", NewIoMod(br, &w).Loader)

			script := fmt.Sprintf(`
      local io = require('io')
      io.write(%s)
      return nil
      `, tc.output)

			err := L.DoString(script)
			assert.NoError(t, err)
			assert.Greater(t, L.GetTop(), 0, "expect result")

			res := lua_convert.FromLuaValue(L.Get(-1))
			assert.Nil(t, res)

			assert.Equal(t, tc.expected, w.Bytes())
		})
	}
}
