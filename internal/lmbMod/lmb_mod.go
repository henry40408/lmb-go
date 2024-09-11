package lmbMod

import (
	"sync"

	"github.com/henry40408/lmb/internal/luaConvert"
	lua "github.com/yuin/gopher-lua"
)

type lmbModule struct {
	// state is scoped to a single evaluation cycle. Different evaluations should not
	// share state. If data sharing across evaluations is required, use the store instead.
	// Example use case for state: HTTP request context
	state *sync.Map
}

func NewLmbModule(state *sync.Map) *lmbModule {
	return &lmbModule{state}
}

func (m *lmbModule) Loader(L *lua.LState) int {
	mod := L.NewTable()

	stateTable := L.NewTable()
	mt := L.NewTable()
	L.SetField(mt, "__index", L.NewFunction(m.get))
	L.SetField(mt, "__newindex", L.NewFunction(m.set))

	L.SetMetatable(stateTable, mt)

	L.SetField(mod, "state", stateTable)

	L.Push(mod)
	return 1
}

func (m *lmbModule) get(L *lua.LState) int {
	key := L.ToString(2)
	raw, ok := m.state.Load(key)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(luaConvert.ToLuaValue(L, raw))
	return 1
}

func (m *lmbModule) set(L *lua.LState) int {
	key := L.ToString(2)
	value := L.Get(3)
	m.state.Store(key, luaConvert.FromLuaValue(value))
	return 0
}
