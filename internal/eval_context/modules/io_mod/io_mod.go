package io_mod

import (
	"bufio"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/henry40408/lmb/internal/lua_convert"
	lua "github.com/yuin/gopher-lua"
)

type ioModule struct {
	reader *bufio.Reader
	writer io.Writer
}

func NewIoMod(br *bufio.Reader, w io.Writer) *ioModule {
	return &ioModule{br, w}
}

func (m *ioModule) Loader(L *lua.LState) int {
	mod := L.NewTable()

	L.SetField(mod, "read", L.NewFunction(m.read))
	L.SetField(mod, "write", L.NewFunction(m.write))

	L.Push(mod)
	return 1
}

func (m *ioModule) read(L *lua.LState) int {
	format := L.Get(1)
	switch v := format.(type) {
	case lua.LNumber:
		return m.readNBytes(L, int(v))
	case lua.LString:
		return m.readWithFormat(L, string(v))
	default:
		L.ArgError(1, "unsupported format")
		return 0
	}
}

func (m *ioModule) readNBytes(L *lua.LState, n int) int {
	buf := make([]byte, n)
	n, err := m.reader.Read(buf)
	if err != nil {
		return m.handleError(L, err)
	}
	buf = buf[:n]
	if utf8.Valid(buf) {
		L.Push(lua_convert.ToLuaValue(L, string(buf)))
	} else {
		L.Push(lua_convert.ToLuaValue(L, buf))
	}
	return 1
}

func (m *ioModule) readWithFormat(L *lua.LState, format string) int {
	switch format {
	case "*a":
		return m.readAll(L)
	case "*n":
		return m.readNumber(L)
	case "*l":
		return m.readLine(L, false)
	case "*L":
		return m.readLine(L, true)
	default:
		L.ArgError(1, "unsupported string format")
		return 0
	}
}

func (m *ioModule) readAll(L *lua.LState) int {
	content, err := io.ReadAll(m.reader)
	if err != nil {
		return m.handleError(L, err)
	}
	L.Push(lua.LString(content))
	return 1
}

func (m *ioModule) readNumber(L *lua.LState) int {
	line, err := m.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return m.handleError(L, err)
	}
	n, err := strconv.ParseFloat(strings.TrimSpace(line), 64)
	if err != nil {
		return m.handleError(L, err)
	}
	L.Push(lua.LNumber(n))
	return 1
}

func (m *ioModule) readLine(L *lua.LState, keepEOL bool) int {
	line, err := m.reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return m.handleError(L, err)
	}
	if !keepEOL {
		line = strings.TrimRight(line, "\n")
	}
	L.Push(lua.LString(line))
	return 1
}

func (m *ioModule) handleError(L *lua.LState, err error) int {
	if err == io.EOF {
		L.Push(lua.LNil)
		return 1
	}
	L.RaiseError(err.Error())
	return 0
}

func (m *ioModule) write(L *lua.LState) int {
	arg := L.Get(1)
	switch v := arg.(type) {
	case lua.LString:
		_, err := io.WriteString(m.writer, string(v))
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	case lua.LNumber:
		_, err := m.writer.Write([]byte(strconv.FormatFloat(float64(v), 'f', -1, 64)))
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	}
	L.ArgError(1, "expect string or number")
	return 0
}
