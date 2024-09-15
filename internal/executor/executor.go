package executor

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	httpMod "github.com/cjoudrey/gluahttp"
	urlMod "github.com/cjoudrey/gluaurl"
	logMod "github.com/cosmotek/loguago"
	"github.com/henry40408/lmb/internal/io_mod"
	"github.com/henry40408/lmb/internal/lmb_mod"
	"github.com/henry40408/lmb/internal/lua_convert"
	"github.com/henry40408/lmb/internal/store"
	"github.com/henry40408/lmb/internal/sync_reader"
	jsonMod "github.com/layeh/gopher-json"
	"github.com/rs/zerolog/log"
	cryptoMod "github.com/tengattack/gluacrypto"
	regexMod "github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

type Executor struct {
	compiled sync.Map
	input    *sync_reader.SyncReader
	store    *store.Store
}

func NewExecutor(store *store.Store, input io.Reader) Executor {
	sr := sync_reader.NewSyncReader(input)
	return Executor{
		compiled: sync.Map{},
		input:    &sr,
		store:    store,
	}
}

func NewTestExecutor(input io.Reader) (Executor, *store.Store) {
	store, err := store.NewStore(":memory:")
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	return NewExecutor(&store, input), &store
}

func (e *Executor) newState(ctx context.Context, state *sync.Map) *lua.LState {
	L := lua.NewState()
	L.SetContext(ctx)
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}

	cryptoMod.Preload(L)

	L.PreloadModule("http", httpMod.NewHttpModule(&http.Client{}).Loader)
	L.PreloadModule("json", jsonMod.Loader)

	logger := logMod.NewLogger(log.Logger)
	L.PreloadModule("logger", logger.Loader)

	L.PreloadModule("re", regexMod.Loader)
	L.PreloadModule("url", urlMod.Loader)

	L.PreloadModule("io", io_mod.NewIoMod(e.input).Loader)
	L.PreloadModule("lmb", lmb_mod.NewLmbModule(state, e.store).Loader)
	return L
}

func (e *Executor) Compile(reader io.Reader, name string) (*lua.FunctionProto, error) {
	start := time.Now()

	parsed, err := parse.Parse(reader, name)
	if err != nil {
		return nil, err
	}
	compiled, err := lua.Compile(parsed, name)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	log.Trace().Str("name", name).Str("duration", duration.String()).Msg("compiled")

	return compiled, nil
}

func (e *Executor) Eval(ctx context.Context, compiled *lua.FunctionProto, state *sync.Map) (interface{}, error) {
	L := e.newState(ctx, state)
	defer L.Close()

	lf := L.NewFunctionFromProto(compiled)
	L.Push(lf)
	if err := L.PCall(0, lua.MultRet, nil); err != nil {
		return nil, err
	}

	if L.GetTop() > 0 {
		result := L.Get(-1)
		L.Pop(1)
		return lua_convert.FromLuaValue(result), nil
	}

	return nil, nil
}

func (e *Executor) findOrCompile(reader io.ReadSeeker) (*lua.FunctionProto, error) {
	hasher := xxhash.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return nil, err
	}

	id := hasher.Sum64()
	if found, ok := e.compiled.Load(id); ok {
		return found.(*lua.FunctionProto), nil
	}

	reader.Seek(0, io.SeekStart)
	compiled, err := e.Compile(reader, strconv.FormatUint(id, 10))
	if err != nil {
		return nil, err
	}

	found, _ := e.compiled.LoadOrStore(id, compiled)
	actual := found.(*lua.FunctionProto)
	return actual, nil
}

func (e *Executor) EvalReader(ctx context.Context, reader io.ReadSeeker, state *sync.Map) (interface{}, error) {
	compiled, err := e.findOrCompile(reader)
	if err != nil {
		return nil, err
	}
	return e.Eval(ctx, compiled, state)
}

func (e *Executor) EvalScript(ctx context.Context, script string, state *sync.Map) (interface{}, error) {
	L := e.newState(ctx, state)
	defer L.Close()

	compiled, err := e.findOrCompile(strings.NewReader(script))
	if err != nil {
		return nil, err
	}

	return e.Eval(ctx, compiled, state)
}
