package lmb_mod

import (
	"sync"

	"github.com/henry40408/lmb/internal/lua_convert"
	"github.com/henry40408/lmb/internal/store"
	lua "github.com/yuin/gopher-lua"
)

type lmbModule struct {
	// state is scoped to a single evaluation cycle. Different evaluations should not
	// share state. If data sharing across evaluations is required, use the store instead.
	// Example use case for state: HTTP request context
	state *sync.Map
	// store represents persistent data storage using SQLite. It's designed to maintain
	// data across multiple evaluation cycles and program executions. Use the store for
	// data that needs to persist long-term and be accessible in future runs.
	store *store.Store
}

func NewLmbModule(state *sync.Map, store *store.Store) *lmbModule {
	return &lmbModule{state, store}
}

func (m *lmbModule) Loader(L *lua.LState) int {
	mod := L.NewTable()

	stateTable := L.NewTable()
	stateMeta := L.NewTable()
	L.SetField(stateMeta, "__index", L.NewFunction(m.get))
	L.SetField(stateMeta, "__newindex", L.NewFunction(m.set))
	L.SetMetatable(stateTable, stateMeta)
	L.SetField(mod, "state", stateTable)

	storeTable := L.NewTable()
	L.SetField(storeTable, "update", L.NewFunction(m.storeUpdate))
	storeMeta := L.NewTable()
	L.SetField(storeMeta, "__index", L.NewFunction(m.storeGet))
	L.SetField(storeMeta, "__newindex", L.NewFunction(m.storePut))
	L.SetMetatable(storeTable, storeMeta)
	L.SetField(mod, "store", storeTable)

	L.Push(mod)
	return 1
}

func (m *lmbModule) get(L *lua.LState) int {
	name := L.CheckString(2)
	value, ok := m.state.Load(name)
	if !ok {
		L.Push(lua.LNil)
		return 1
	}
	L.Push(lua_convert.ToLuaValue(L, value))
	return 1
}

func (m *lmbModule) set(L *lua.LState) int {
	key := L.CheckString(2)
	data := L.Get(3)
	m.state.Store(key, lua_convert.FromLuaValue(data))
	return 0
}

func (m *lmbModule) storeGet(L *lua.LState) int {
	name := L.CheckString(2)
	value, err := m.store.Get(name)
	if err != nil {
		L.RaiseError(err.Error())
	}
	L.Push(lua_convert.ToLuaValue(L, value))
	return 1
}

func (m *lmbModule) storePut(L *lua.LState) int {
	name := L.CheckString(2)
	value := lua_convert.FromLuaValue(L.Get(3))
	err := m.store.Put(name, value)
	if err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}

func (m *lmbModule) storeUpdate(L *lua.LState) int {
	f := L.CheckFunction(2)

	st, err := m.store.Begin()
	if err != nil {
		L.RaiseError(err.Error())
	}
	defer st.Rollback()

	t := L.NewTable()
	mt := L.NewTable()
	L.SetField(mt, "__index", L.NewFunction(func(l *lua.LState) int {
		name := L.CheckString(2)
		value, err := st.Get(name)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua_convert.ToLuaValue(L, value))
		return 1
	}))
	L.SetField(mt, "__newindex", L.NewFunction(func(l *lua.LState) int {
		name := L.CheckString(2)
		value := lua_convert.FromLuaValue(L.Get(3))
		err := st.Put(name, value)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	}))
	L.SetMetatable(t, mt)

	L.Push(f)
	L.Push(t)
	err = L.PCall(1, lua.MultRet, nil)
	if err != nil {
		L.RaiseError(err.Error())
	}

	err = st.Commit()
	if err != nil {
		L.RaiseError(err.Error())
	}

	nResults := L.GetTop() - 1
	if nResults == 0 {
		L.Push(lua.LNil)
		return 1
	}
	return nResults
}
