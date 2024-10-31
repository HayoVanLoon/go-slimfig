package slimfig

import (
	"fmt"
	"reflect"
	"strconv"
)

func toString(a any) string {
	if s, ok := a.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", a)
}

func toString2(a any) (string, bool) {
	return toString(a), true
}

func toInt(a any) (int, bool) {
	// uint and similar might cause overflows
	switch x := a.(type) {
	case int:
		return x, true
	case int8:
		return int(x), true
	case int16:
		return int(x), true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float32:
		return int(x), true
	case float64:
		return int(x), true
	}
	i, err := strconv.Atoi(toString(a))
	if err != nil {
		return 0, false
	}
	return i, true
}

func toFloat64(a any) (float64, bool) {
	switch x := a.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	}
	f, err := strconv.ParseFloat(toString(a), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func toBool(a any) (bool, bool) {
	if b, ok := a.(bool); ok {
		return b, true
	}
	b, err := strconv.ParseBool(fmt.Sprintf("%v", a))
	if err != nil {
		return false, false
	}
	return b, true
}

func toSlice[T any](a any, conv func(a any) (T, bool)) ([]T, bool) {
	if out, ok := a.([]T); ok {
		return out, true
	}
	v := reflect.ValueOf(a)
	if v.Type().Kind() != reflect.Slice {
		return nil, false
	}
	if v.Len() == 0 {
		return nil, true
	}
	var out []T
	v.Seq()(func(v reflect.Value) bool {
		if v2, ok := conv(v.Interface()); ok {
			out = append(out, v2)
		}
		return true
	})
	return out, true
}

func toMap[T any](a any, conv func(a any) (T, bool)) (map[string]T, bool) {
	if out, ok := a.(map[string]T); ok {
		return out, true
	}
	v := reflect.ValueOf(a)
	if v.Type().Kind() != reflect.Map {
		return nil, false
	}
	if v.Len() == 0 {
		return nil, true
	}
	iter := v.MapRange()
	out := make(map[string]T)
	for iter.Next() {
		k := toString(iter.Key().Interface())
		if v2, ok := conv(iter.Value().Interface()); ok {
			out[k] = v2
		}
	}
	return out, true
}
