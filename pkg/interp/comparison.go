package interp

import (
	"strconv"
	"strings"
)

// evalComparison checks for a top-level comparison operator (<=, >=, ==, !=, <, >)
// and if found, evaluates both sides and returns the boolean result.
// Returns (value, found, err). found=false means no comparison operator at this level,
// caller should fall through to other evaluation strategies.
func (i *Interpreter) evalComparison(expr string) (Value, bool, error) {
	depth := 0
	// scan left to right, find first top-level comparison op (comparisons are non-associative,
	// so first occurrence is fine for our simple expr subset)
	for idx := 0; idx < len(expr); idx++ {
		ch := expr[idx]
		switch ch {
		case '(', '[', '{':
			depth++
			continue
		case ')', ']', '}':
			depth--
			continue
		}
		if depth != 0 {
			continue
		}

		var opLen int
		var op string
		switch {
		case idx+1 < len(expr) && ch == '<' && expr[idx+1] == '=':
			op, opLen = "<=", 2
		case idx+1 < len(expr) && ch == '>' && expr[idx+1] == '=':
			op, opLen = ">=", 2
		case idx+2 < len(expr) && ch == '=' && expr[idx+1] == '=' && expr[idx+2] == '=':
			op, opLen = "===", 3
		case idx+1 < len(expr) && ch == '=' && expr[idx+1] == '=':
			op, opLen = "==", 2
		case idx+1 < len(expr) && ch == '!' && expr[idx+1] == '=':
			op, opLen = "!=", 2
		case ch == '<':
			op, opLen = "<", 1
		case ch == '>':
			op, opLen = ">", 1
		default:
			continue
		}

		leftStr := strings.TrimSpace(expr[:idx])
		rightStr := strings.TrimSpace(expr[idx+opLen:])
		if leftStr == "" || rightStr == "" {
			continue
		}

		leftVal, err := i.evalExpr(leftStr)
		if err != nil {
			return nil, true, err
		}
		rightVal, err := i.evalExpr(rightStr)
		if err != nil {
			return nil, true, err
		}

		result := compareValues(leftVal, rightVal, op)
		return result, true, nil
	}

	return nil, false, nil
}

// compareValues compares two runtime values using the given operator.
// Numbers compare numerically; falls back to string comparison otherwise.
func compareValues(left, right Value, op string) bool {
	lf, lok := toFloat(left)
	rf, rok := toFloat(right)

	if lok && rok {
		switch op {
		case "<":
			return lf < rf
		case ">":
			return lf > rf
		case "<=":
			return lf <= rf
		case ">=":
			return lf >= rf
		case "==", "===":
			return lf == rf
		case "!=":
			return lf != rf
		}
	}

	ls := toStr(left)
	rs := toStr(right)
	switch op {
	case "<":
		return ls < rs
	case ">":
		return ls > rs
	case "<=":
		return ls <= rs
	case ">=":
		return ls >= rs
	case "==", "===":
		return ls == rs
	case "!=":
		return ls != rs
	}
	return false
}

func toFloat(v Value) (float64, bool) {
	switch x := v.(type) {
	case int64:
		return float64(x), true
	case int:
		return float64(x), true
	case float64:
		return x, true
	case string:
		f, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	}
	return 0, false
}

func toStr(v Value) string {
	switch x := v.(type) {
	case string:
		return x
	case int64:
		return strconv.FormatInt(x, 10)
	case int:
		return strconv.Itoa(x)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	}
	return ""
}
