package internal

import (
	"github.com/yuin/gopher-lua"
)

type Executor struct {
	Script string
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

func (e *Executor) Eval() (interface{}, error) {
	L := lua.NewState()
	defer L.Close()

	if err := L.DoString(e.Script); err != nil {
		return nil, err
	}

	if L.GetTop() > 0 {
		result := L.Get(-1)
		L.Pop(1)
		return fromLuaValue(result), nil
	}

	return nil, nil
}
