package internal

import (
	"io"
	"net/http"
	"os"

	httpMod "github.com/cjoudrey/gluahttp"
	urlMod "github.com/cjoudrey/gluaurl"
	logMod "github.com/cosmotek/loguago"
	jsonMod "github.com/layeh/gopher-json"
	"github.com/rs/zerolog"
	cryptoMod "github.com/tengattack/gluacrypto"
	regexMod "github.com/yuin/gluare"
	lua "github.com/yuin/gopher-lua"
)

type Executor struct {
	L *lua.LState
}

func NewExecutor() Executor {
	L := lua.NewState()

	cryptoMod.Preload(L)

	L.PreloadModule("http", httpMod.NewHttpModule(&http.Client{}).Loader)
	L.PreloadModule("json", jsonMod.Loader)

	zlogger := zerolog.New(os.Stdout).With().Logger()
	logger := logMod.NewLogger(zlogger)
	L.PreloadModule("logger", logger.Loader)

	L.PreloadModule("re", regexMod.Loader)
	L.PreloadModule("url", urlMod.Loader)
	return Executor{L: L}
}

func (e *Executor) Close() {
	e.L.Close()
}

func fromLuaValue(lv lua.LValue) interface{} {
	switch v := lv.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(v)
	case lua.LNumber:
		return float64(v)
	case lua.LString:
		return string(v)
	case *lua.LTable:
		maxn := v.MaxN()
		if maxn == 0 { // table
			ret := make(map[string]interface{})
			v.ForEach(func(key, value lua.LValue) {
				ret[key.String()] = fromLuaValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, fromLuaValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v.String()
	}
}

func (e *Executor) Eval(script string) (interface{}, error) {
	if err := e.L.DoString(script); err != nil {
		return nil, err
	}

	if e.L.GetTop() > 0 {
		result := e.L.Get(-1)
		e.L.Pop(1)
		return fromLuaValue(result), nil
	}

	return nil, nil
}

func (e *Executor) EvalFile(filePath string) (interface{}, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	script, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return e.Eval(string(script))
}
