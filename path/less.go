package path

import "time"

import "cloud.google.com/go/civil"

func Less(obj any, path string, v1Str string) (bool, error) {
	v2, err := getAny(obj, path)
	if err != nil {
		return false, err
	}
	if v2 == nil {
		return false, nil
	}

	v1, err := parse(v1Str, v2)
	if err != nil {
		return false, err
	}

	return less(v1, v2), nil
}

func less(v1, v2 any) bool {
	if v2 == nil {
		return false
	}

	if isSlice(v2) {
		return anyTrue(v2, func(x any) bool { return less(v1, x) })
	}

	switch v2t := v2.(type) {
	case int:
		return v2t < v1.(int)

	case int64:
		return v2t < v1.(int64)

	case uint:
		return v2t < v1.(uint)

	case uint64:
		return v2t < v1.(uint64)

	case float32:
		return v2t < v1.(float32)

	case float64:
		return v2t < v1.(float64)

	case string:
		return v2t < v1.(string)

	case bool:
		return v2t == false && v1.(bool) == true

	case time.Time:
		tm := v1.(*timeVal)
		return v2t.Truncate(tm.precision).Before(tm.time)

	case civil.Date:
		return v2t.Before(v1.(civil.Date))

	default:
		panic(v2)
	}
}
