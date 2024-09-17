package lmb_mod

import (
	"strings"
	"sync"
	"testing"

	"github.com/henry40408/lmb/internal/eval_context/testutil"
	"github.com/henry40408/lmb/internal/lua_convert"
	"github.com/henry40408/lmb/internal/store"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

func TestStore(t *testing.T) {
	L, _, store := setupEvalContext()
	defer store.Close()
	defer L.Close()

	err := L.DoString(`
  local m = require('lmb')
  m.store['a'] = 47
  assert(m.store['a'] == 47)
  assert(not m.store['b'])
  return true
  `)
	assert.NoError(t, err)

	res := lua_convert.FromLuaValue(L.Get(-1))
	assert.Equal(t, true, res)
}

func TestStoreUpdate(t *testing.T) {
	L, _, store := setupEvalContext()
	defer store.Close()
	defer L.Close()

	err := L.DoString(`
  local m = require('lmb')
  m.store['alice'] = 50
  m.store['bob'] = 50
  m.store:update(function(store)
    local alice = store['alice']
    if alice < 100 then
      error('insufficient fund')
    end
    store['alice'] = store['alice'] - 100
    store['bob'] = store['bob'] + 100
  end)
  return true
  `)
	assert.Error(t, err, "insufficient fund")

	failed := lua_convert.FromLuaValue(L.Get(-1))
	assert.Nil(t, failed)

	alice, err := store.Get("alice")
	assert.NoError(t, err)
	assert.Equal(t, 50.0, alice)
	bob, err := store.Get("bob")
	assert.NoError(t, err)
	assert.Equal(t, 50.0, bob)

	err = L.DoString(`
  local m = require('lmb')
  m.store['alice'] = 100
  m.store['bob'] = 0
  m.store:update(function(store)
    local alice = store['alice']
    if alice < 100 then
      error('insufficient fund')
    end
    store['alice'] = store['alice'] - 100
    store['bob'] = store['bob'] + 100
  end)
  return true
  `)
	assert.NoError(t, err)

	success := lua_convert.FromLuaValue(L.Get(-1))
	assert.Equal(t, true, success)

	alice, err = store.Get("alice")
	assert.NoError(t, err)
	assert.Equal(t, 0.0, alice)
	bob, err = store.Get("bob")
	assert.NoError(t, err)
	assert.Equal(t, 100.0, bob)
}

func TestStoreUpdateConcurrency(t *testing.T) {
	store, err := store.NewStore(":memory:")
	assert.NoError(t, err)
	defer store.Close()

	reader := strings.NewReader(`
  local m = require('lmb')
  m.store:update(function(store)
    store['counter'] = (store['counter'] or 0) + 1
  end)
  return true
  `)
	chunk, err := parse.Parse(reader, "compiled")
	assert.NoError(t, err)
	proto, err := lua.Compile(chunk, "compiled")
	assert.NoError(t, err)

	var wg sync.WaitGroup

	count := 100
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			var state sync.Map

			defer wg.Done()

			L := testutil.NewLuaTestState()
			defer L.Close()

			L.PreloadModule("lmb", NewLmbModule(&state, store).Loader)

			L.Push(L.NewFunctionFromProto(proto))
			err := L.PCall(0, lua.MultRet, nil)
			assert.NoError(t, err)

			res := lua_convert.FromLuaValue(L.Get(-1))
			assert.Equal(t, true, res)
		}(i)
	}

	wg.Wait()

	value, err := store.Get("counter")
	assert.NoError(t, err)
	assert.Equal(t, float64(count), value)
}

func setupEvalContext() (*lua.LState, *sync.Map, *store.Store) {
	L := testutil.NewLuaTestState()

	var state sync.Map
	state.Store("a", 1.0)

	store, err := store.NewStore(":memory:")
	if err != nil {
		panic(err)
	}
	L.PreloadModule("lmb", NewLmbModule(&state, store).Loader)

	return L, &state, store
}
