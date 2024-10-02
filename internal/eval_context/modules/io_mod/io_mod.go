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
		// read N bytes
		buf := make([]byte, uint(v))
		n, err := m.reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				L.Push(lua.LNil)
				return 1
			}
			L.RaiseError(err.Error())
		}
		buf = buf[:n]
		if utf8.Valid(buf) {
			L.Push(lua_convert.ToLuaValue(L, string(buf)))
		} else {
			L.Push(lua_convert.ToLuaValue(L, buf))
		}
		return 1
	case lua.LString:
		switch string(v) {
		case "*a":
			// "*a" # Reads the whole file.
			content, err := io.ReadAll(m.reader)
			if err != nil {
				if err == io.EOF {
					L.Push(lua.LNil)
					return 1
				}
				L.RaiseError(err.Error())
			}
			L.Push(lua_convert.ToLuaValue(L, string(content)))
			return 1
		case "*n":
			// "*n" # Reads a numeral and returns it as number.
			line, err := m.reader.ReadString('\n')
			if err != nil && err != io.EOF {
				L.RaiseError(err.Error())
			}
			n, err := strconv.ParseFloat(string(line), 64)
			if err != nil {
				L.RaiseError(err.Error())
			}
			L.Push(lua_convert.ToLuaValue(L, n))
			return 1
		case "*l":
			// "*l" # Reads the next line skipping the end of line.
			line, err := m.reader.ReadString('\n')
			if err != nil && err != io.EOF {
				L.RaiseError(err.Error())
			}
			L.Push(lua_convert.ToLuaValue(L, strings.TrimRight(line, "\n")))
			return 1
		case "*L":
			// "*L" # Reads the next line keeping the end of line.
			line, err := m.reader.ReadString('\n')
			if err != nil && err != io.EOF {
				L.RaiseError(err.Error())
			}
			L.Push(lua_convert.ToLuaValue(L, line))
			return 1
		}
	}
	L.ArgError(1, "unsupported format")
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
