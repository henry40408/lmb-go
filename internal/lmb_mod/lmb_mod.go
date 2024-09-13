package lmb_mod

import (
	"bytes"
	"database/sql"
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/henry40408/lmb/internal/lua_convert"
	lua "github.com/yuin/gopher-lua"
)

type lmbModule struct {
	// state is scoped to a single evaluation cycle. Different evaluations should not
	// share state. If data sharing across evaluations is required, use the store instead.
	// Example use case for state: HTTP request context
	state *sync.Map
	store *sql.DB
}

func NewLmbModule(state *sync.Map, store *sql.DB) *lmbModule {
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
	storeMeta := L.NewTable()
	L.SetField(storeMeta, "__index", L.NewFunction(m.storeGet))
	L.SetField(storeMeta, "__newindex", L.NewFunction(m.storePut))
	L.SetMetatable(storeTable, storeMeta)
	L.SetField(mod, "store", storeTable)

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
	L.Push(lua_convert.ToLuaValue(L, raw))
	return 1
}

func (m *lmbModule) set(L *lua.LState) int {
	key := L.ToString(2)
	value := L.Get(3)
	m.state.Store(key, lua_convert.FromLuaValue(value))
	return 0
}

func serializeData(data interface{}) []byte {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	encoder.Encode(data)
	return buffer.Bytes()
}

func deserializeData(value []byte, target interface{}) error {
	decoder := gob.NewDecoder(bytes.NewBuffer(value))
	return decoder.Decode(target)
}

func (m *lmbModule) storeGet(L *lua.LState) int {
	name := L.ToString(2)
	stmt, err := m.store.Prepare(`SELECT value FROM store WHERE name = ?`)
	if err != nil {
		L.RaiseError(err.Error())
	}
	var value []byte
	err = stmt.QueryRow(&name).Scan(&value)
	fmt.Printf("%v", value)
	if err != nil {
		if err == sql.ErrNoRows {
			L.Push(lua.LNil)
			return 1
		} else {
			L.RaiseError(err.Error())
		}
	}
	var deserialized interface{}
	err = deserializeData(value, &deserialized)
	fmt.Printf("%v", deserialized)
	if err != nil {
		L.RaiseError(err.Error())
	}
	fmt.Printf("%v", deserialized)
	L.Push(lua_convert.ToLuaValue(L, deserialized))
	return 1
}

func (m *lmbModule) storePut(L *lua.LState) int {
	name := L.ToString(2)
	value := L.Get(3)
	data := lua_convert.FromLuaValue(value)
	stmt, err := m.store.Prepare(`INSERT INTO store (name, value, type_hint, size) VALUES (?, ?, ?, ?)`)
	if err != nil {
		L.RaiseError(err.Error())
	}
	serialized := serializeData(&data)
	_, err = stmt.Exec(&name, serialized, reflect.TypeOf(data).Name(), int64(unsafe.Sizeof(data)))
	if err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}
