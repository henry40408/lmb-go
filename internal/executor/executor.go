package executor

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	httpMod "github.com/cjoudrey/gluahttp"
	urlMod "github.com/cjoudrey/gluaurl"
	logMod "github.com/cosmotek/loguago"
	"github.com/henry40408/lmb/internal/database"
	"github.com/henry40408/lmb/internal/lmb_mod"
	"github.com/henry40408/lmb/internal/lua_convert"
	jsonMod "github.com/layeh/gopher-json"
	"github.com/rs/zerolog/log"
	cryptoMod "github.com/tengattack/gluacrypto"
	regexMod "github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
)

type Executor struct {
	compiled sync.Map
	store    *sql.DB
}

func NewExecutor(store *sql.DB) Executor {
	return Executor{
		compiled: sync.Map{},
		store:    store,
	}
}

func NewTestExecutor() (Executor, *sql.DB) {
	db, err := database.OpenDB(":memory:")
	if err != nil {
		log.Fatal().Err(err)
	}
	return NewExecutor(db), db
}

func (e *Executor) newState(ctx context.Context, state *sync.Map, store *sql.DB) *lua.LState {
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

	logger := logMod.NewLogger(log.With().Logger())
	L.PreloadModule("logger", logger.Loader)

	L.PreloadModule("re", regexMod.Loader)
	L.PreloadModule("url", urlMod.Loader)

	L.PreloadModule("lmb", lmb_mod.NewLmbModule(state, store).Loader)
	return L
}

func (e *Executor) Compile(reader io.Reader, hash string) (*lua.FunctionProto, error) {
	logger := log.With().Str("hash", hash).Logger()
	start := time.Now()

	parsed, err := parse.Parse(reader, hash)
	if err != nil {
		return nil, err
	}
	compiled, err := lua.Compile(parsed, hash)
	if err != nil {
		return nil, err
	}

	duration := time.Since(start)
	logger.Trace().Dur("duration", duration).Msg("compiled")

	return compiled, nil
}

func (e *Executor) Eval(ctx context.Context, compiled *lua.FunctionProto, state *sync.Map, store *sql.DB) (interface{}, error) {
	L := e.newState(ctx, state, store)
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

func (e *Executor) EvalFile(ctx context.Context, filePath string, state *sync.Map, store *sql.DB) (interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	compiled, err := e.findOrCompile(file)
	if err != nil {
		return nil, err
	}

	return e.Eval(ctx, compiled, state, store)
}

func (e *Executor) EvalScript(ctx context.Context, script string, state *sync.Map, store *sql.DB) (interface{}, error) {
	L := e.newState(ctx, state, store)
	defer L.Close()

	compiled, err := e.findOrCompile(strings.NewReader(script))
	if err != nil {
		return nil, err
	}

	return e.Eval(ctx, compiled, state, store)
}
