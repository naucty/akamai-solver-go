package interp2

import (
	"fmt"
	"math"
	"strconv"
)

func toFloat64(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func toFloat64Val(v Value) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	case bool:
		if x {
			return 1
		}
		return 0
	case nil:
		return 0
	case []Value:
		if len(x) == 0 {
			return 0
		}
		return toFloat64Val(x[0])
	default:
		return math.NaN()
	}
}

func toInt64Val(v Value) int64 {
	switch x := v.(type) {
	case float64:
		return int64(x)
	case string:
		i, _ := strconv.ParseInt(x, 10, 64)
		return i
	case bool:
		if x {
			return 1
		}
		return 0
	case nil:
		return 0
	default:
		return 0
	}
}

func toStringVal(v Value) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		return formatNumber(x)
	case bool:
		if x {
			return "true"
		}
		return "false"
	case nil:
		return "undefined"
	case []Value:
		return fmt.Sprintf("[object Array]")
	default:
		return "[object Object]"
	}
}

func toKeyString(v Value) string {
	return toStringVal(v)
}

func formatNumber(f float64) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 1) {
		return "Infinity"
	}
	if math.IsInf(f, -1) {
		return "-Infinity"
	}
	if f == float64(int64(f)) {
		return strconv.FormatInt(int64(f), 10)
	}
	return strconv.FormatFloat(f, 'g', -1, 64)
}

func looseEquals(a, b Value) bool {
	if strictEquals(a, b) {
		return true
	}
	if a == nil || b == nil {
		return a == b
	}
	af, aok := a.(float64)
	bf, bok := b.(float64)
	if aok && bok {
		return af == bf || (math.IsNaN(af) && math.IsNaN(bf))
	}
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return as == bs
	}
	// bool == number coercion
	if ab, ok := a.(bool); ok {
		return looseEquals(float64Cond(ab), b)
	}
	if bb, ok := b.(bool); ok {
		return looseEquals(a, float64Cond(bb))
	}
	return false
}

func float64Cond(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

func strictEquals(a, b Value) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	switch a := a.(type) {
	case float64:
		if b, ok := b.(float64); ok {
			if math.IsNaN(a) && math.IsNaN(b) {
				return false
			}
			return a == b
		}
	case string:
		if b, ok := b.(string); ok {
			return a == b
		}
	case bool:
		if b, ok := b.(bool); ok {
			return a == b
		}
	case *Object:
		return a == b
	case *Function:
		return a == b
	}
	return false
}

func compareOp(a, b Value, op string) bool {
	af := toFloat64Val(a)
	bf := toFloat64Val(b)
	switch op {
	case "<":
		return af < bf
	case ">":
		return af > bf
	case "<=":
		return af <= bf
	case ">=":
		return af >= bf
	}
	return false
}
