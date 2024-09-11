package luaConvert

import (
	"reflect"

	lua "github.com/yuin/gopher-lua"
)

func FromLuaValue(lv lua.LValue) interface{} {
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
				ret[key.String()] = FromLuaValue(value)
			})
			return ret
		} else { // array
			ret := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				ret = append(ret, FromLuaValue(v.RawGetInt(i)))
			}
			return ret
		}
	default:
		return v.String()
	}
}

func ToLuaValue[T any](L *lua.LState, value T) lua.LValue {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Bool:
		return lua.LBool(v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(v.Uint())
	case reflect.Float32, reflect.Float64:
		return lua.LNumber(v.Float())
	case reflect.String:
		return lua.LString(v.String())
	case reflect.Slice, reflect.Array:
		table := L.CreateTable(v.Len(), 0)
		for i := 0; i < v.Len(); i++ {
			table.Append(ToLuaValue(L, v.Index(i).Interface()))
		}
		return table
	case reflect.Map:
		table := L.CreateTable(0, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			table.RawSet(ToLuaValue(L, iter.Key().Interface()), ToLuaValue(L, iter.Value().Interface()))
		}
		return table
	case reflect.Struct:
		table := L.CreateTable(0, v.NumField())
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			table.RawSet(lua.LString(field.Name), ToLuaValue(L, v.Field(i).Interface()))
		}
		return table
	case reflect.Ptr:
		if v.IsNil() {
			return lua.LNil
		}
		return ToLuaValue(L, v.Elem().Interface())
	default:
		return lua.LNil
	}
}
