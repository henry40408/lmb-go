package io_mod

import (
	"io"
	"strconv"
	"unicode/utf8"

	"github.com/henry40408/lmb/internal/lua_convert"
	"github.com/henry40408/lmb/internal/sync_reader"
	lua "github.com/yuin/gopher-lua"
)

type ioModule struct {
	sr *sync_reader.SyncReader
}

func NewIoMod(sr *sync_reader.SyncReader) *ioModule {
	return &ioModule{sr}
}

func (m *ioModule) Loader(L *lua.LState) int {
	mod := L.NewTable()

	L.SetField(mod, "read", L.NewFunction(m.read))

	L.Push(mod)
	return 1
}

func (m *ioModule) read(L *lua.LState) int {
	format := L.Get(1)
	switch v := format.(type) {
	case lua.LNumber:
		// read N bytes
		buf := make([]byte, uint(v))
		n, err := m.sr.Read(buf)
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
			content, err := m.sr.ReadAll()
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
			var line []byte
			for {
				fragment, isPrefix, err := m.sr.ReadLine()
				if err != nil {
					if err == io.EOF {
						L.Push(lua.LNil)
						return 1
					}
					L.RaiseError(err.Error())
				}
				line = append(line, fragment...)
				if !isPrefix {
					break
				}
			}
			n, err := strconv.ParseFloat(string(line), 64)
			if err != nil {
				L.RaiseError(err.Error())
			}
			L.Push(lua_convert.ToLuaValue(L, n))
			return 1
		case "*l":
			// "*l" # Reads the next line skipping the end of line.
			var line []byte
			for {
				fragment, isPrefix, err := m.sr.ReadLine()
				if err != nil {
					if err == io.EOF {
						L.Push(lua.LNil)
						return 1
					}
					L.RaiseError(err.Error())
				}
				line = append(line, fragment...)
				if !isPrefix {
					break
				}
			}
			L.Push(lua_convert.ToLuaValue(L, string(line)))
			return 1
		case "*L":
			// "*L" # Reads the next line keeping the end of line.
			var line []byte
			for {
				fragment, isPrefix, err := m.sr.ReadLine()
				if err != nil {
					if err == io.EOF {
						L.Push(lua.LNil)
						return 1
					}
					L.RaiseError(err.Error())
				}
				line = append(line, fragment...)
				if !isPrefix {
					break
				}
			}
			line = append(line, '\n')
			L.Push(lua_convert.ToLuaValue(L, string(line)))
			return 1
		}
	}
	L.RaiseError("unsupported format %v", format)
	return 0
}
