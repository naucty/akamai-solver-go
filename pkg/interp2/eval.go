package interp2

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/dop251/goja/ast"
)

// evalExpr evaluates an AST expression node to a runtime Value
func (interp *Interpreter) evalExpr(expr ast.Expression, scope *Scope) (Value, error) {
	switch e := expr.(type) {
	case nil:
		return nil, nil

	case *ast.NumberLiteral:
		switch v := e.Value.(type) {
		case float64:
			return v, nil
		case int64:
			return float64(v), nil
		case string:
			return toFloat64(v), nil
		}
		return 0.0, nil

	case *ast.StringLiteral:
		return e.Value.String(), nil

	case *ast.BooleanLiteral:
		return e.Value, nil

	case *ast.NullLiteral:
		return nil, nil

	case *ast.Identifier:
		name := e.Name.String()
		if v, ok := scope.Get(name); ok {
			return v, nil
		}
		return nil, fmt.Errorf("undefined variable: %s", name)

	case *ast.ArrayLiteral:
		arr := make([]Value, 0, len(e.Value))
		for _, el := range e.Value {
			if el == nil {
				arr = append(arr, nil)
				continue
			}
			v, err := interp.evalExpr(el, scope)
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil

	case *ast.ObjectLiteral:
		obj := &Object{Props: make(map[string]Value)}
		for _, prop := range e.Value {
			if p, ok := prop.(*ast.PropertyKeyed); ok {
				key := interp.propKeyName(p.Key, scope)
				val, err := interp.evalExpr(p.Value, scope)
				if err != nil {
					return nil, err
				}
				obj.Props[key] = val
			}
		}
		return obj, nil

	case *ast.AssignExpression:
		return interp.evalAssign(e, scope)

	case *ast.BinaryExpression:
		return interp.evalBinary(e, scope)

	case *ast.UnaryExpression:
		return interp.evalUnary(e, scope)

	case *ast.ConditionalExpression:
		cond, err := interp.evalExpr(e.Test, scope)
		if err != nil {
			return nil, err
		}
		if isTruthy(cond) {
			return interp.evalExpr(e.Consequent, scope)
		}
		return interp.evalExpr(e.Alternate, scope)

	case *ast.SequenceExpression:
		var last Value
		for _, sub := range e.Sequence {
			v, err := interp.evalExpr(sub, scope)
			if err != nil {
				return nil, err
			}
			last = v
		}
		return last, nil

	case *ast.CallExpression:
		return interp.evalCall(e, scope)

	case *ast.BracketExpression:
		obj, err := interp.evalExpr(e.Left, scope)
		if err != nil {
			return nil, err
		}
		keyVal, err := interp.evalExpr(e.Member, scope)
		if err != nil {
			return nil, err
		}
		return interp.getMember(obj, toKeyString(keyVal))

	case *ast.DotExpression:
		obj, err := interp.evalExpr(e.Left, scope)
		if err != nil {
			return nil, err
		}
		return interp.getMember(obj, e.Identifier.Name.String())

	case *ast.FunctionLiteral:
		return interp.makeFunction(e, scope), nil

	default:
		return nil, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

func (interp *Interpreter) propKeyName(key ast.Expression, scope *Scope) string {
	switch k := key.(type) {
	case *ast.Identifier:
		return k.Name.String()
	case *ast.StringLiteral:
		return k.Value.String()
	case *ast.NumberLiteral:
		switch v := k.Value.(type) {
		case float64:
			return formatNumber(v)
		case int64:
			return formatNumber(float64(v))
		}
		return ""
	}
	return ""
}

func (interp *Interpreter) getMember(container Value, key string) (Value, error) {
	switch c := container.(type) {
	case []Value:
		if key == "length" {
			return float64(len(c)), nil
		}
		idx, err := strconv.Atoi(key)
		if err != nil {
			return nil, nil
		}
		if idx < 0 || idx >= len(c) {
			return nil, nil
		}
		return c[idx], nil
	case string:
		if key == "length" {
			return float64(len(c)), nil
		}
		idx, err := strconv.Atoi(key)
		if err == nil {
			if idx < 0 || idx >= len(c) {
				return "", nil
			}
			return string(c[idx]), nil
		}
		return nativeStringMethod(c, key)
	case *Object:
		if v, ok := c.Props[key]; ok {
			return v, nil
		}
		return nil, nil
	case nil:
		return nil, fmt.Errorf("cannot read property '%s' of null/undefined", key)
	}
	return nil, nil
}

func (interp *Interpreter) setMember(container Value, key string, val Value) error {
	switch c := container.(type) {
	case *Object:
		c.Props[key] = val
		return nil
	}
	return fmt.Errorf("cannot set property on %T", container)
}

func (interp *Interpreter) evalAssign(e *ast.AssignExpression, scope *Scope) (Value, error) {
	rhs, err := interp.evalExpr(e.Right, scope)
	if err != nil {
		return nil, err
	}

	op := e.Operator.String()
	if op != "=" {
		lhsVal, err := interp.evalExpr(e.Left, scope)
		if err != nil {
			return nil, err
		}
		rhs, err = applyCompoundOp(op, lhsVal, rhs)
		if err != nil {
			return nil, err
		}
	}

	switch lhs := e.Left.(type) {
	case *ast.Identifier:
		scope.Set(lhs.Name.String(), rhs)
	case *ast.BracketExpression:
		objVal, err := interp.evalExpr(lhs.Left, scope)
		if err != nil {
			return nil, err
		}
		keyVal, err := interp.evalExpr(lhs.Member, scope)
		if err != nil {
			return nil, err
		}
		if err := interp.setMember(objVal, toKeyString(keyVal), rhs); err != nil {
			return nil, err
		}
	case *ast.DotExpression:
		objVal, err := interp.evalExpr(lhs.Left, scope)
		if err != nil {
			return nil, err
		}
		if err := interp.setMember(objVal, lhs.Identifier.Name.String(), rhs); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported assignment target: %T", e.Left)
	}
	return rhs, nil
}

func applyCompoundOp(op string, left, right Value) (Value, error) {
	base := strings.TrimSuffix(op, "=")
	return binaryOp(base, left, right)
}

func (interp *Interpreter) evalBinary(e *ast.BinaryExpression, scope *Scope) (Value, error) {
	left, err := interp.evalExpr(e.Left, scope)
	if err != nil {
		return nil, err
	}
	op := e.Operator.String()

	if op == "&&" {
		if !isTruthy(left) {
			return left, nil
		}
		return interp.evalExpr(e.Right, scope)
	}
	if op == "||" {
		if isTruthy(left) {
			return left, nil
		}
		return interp.evalExpr(e.Right, scope)
	}

	right, err := interp.evalExpr(e.Right, scope)
	if err != nil {
		return nil, err
	}
	return binaryOp(op, left, right)
}

func binaryOp(op string, left, right Value) (Value, error) {
	switch op {
	case "+":
		if ls, ok := left.(string); ok {
			return ls + toStringVal(right), nil
		}
		if rs, ok := right.(string); ok {
			return toStringVal(left) + rs, nil
		}
		return toFloat64Val(left) + toFloat64Val(right), nil
	case "-":
		return toFloat64Val(left) - toFloat64Val(right), nil
	case "*":
		return toFloat64Val(left) * toFloat64Val(right), nil
	case "/":
		return toFloat64Val(left) / toFloat64Val(right), nil
	case "%":
		return math.Mod(toFloat64Val(left), toFloat64Val(right)), nil
	case "&":
		return float64(toInt64Val(left) & toInt64Val(right)), nil
	case "|":
		return float64(toInt64Val(left) | toInt64Val(right)), nil
	case "^":
		return float64(toInt64Val(left) ^ toInt64Val(right)), nil
	case "<<":
		return float64(toInt64Val(left) << uint(toInt64Val(right)&31)), nil
	case ">>":
		return float64(toInt64Val(left) >> uint(toInt64Val(right)&31)), nil
	case ">>>":
		return float64(uint32(toInt64Val(left)) >> uint(toInt64Val(right)&31)), nil
	case "<":
		return compareOp(left, right, "<"), nil
	case ">":
		return compareOp(left, right, ">"), nil
	case "<=":
		return compareOp(left, right, "<="), nil
	case ">=":
		return compareOp(left, right, ">="), nil
	case "==":
		return looseEquals(left, right), nil
	case "===":
		return strictEquals(left, right), nil
	case "!=":
		return !looseEquals(left, right), nil
	case "!==":
		return !strictEquals(left, right), nil
	}
	return nil, fmt.Errorf("unsupported binary operator: %s", op)
}

// evalUnary handles !, -, +, ~, void, typeof, and prefix/postfix ++/--
func (interp *Interpreter) evalUnary(e *ast.UnaryExpression, scope *Scope) (Value, error) {
	opStr := e.Operator.String()

	if opStr == "typeof" {
		v, err := interp.evalExpr(e.Operand, scope)
		if err != nil {
			return "undefined", nil
		}
		return jsTypeOf(v), nil
	}

	if opStr == "++" || opStr == "--" {
		old, err := interp.evalExpr(e.Operand, scope)
		if err != nil {
			return nil, err
		}
		oldF := toFloat64Val(old)
		var newF float64
		if opStr == "++" {
			newF = oldF + 1
		} else {
			newF = oldF - 1
		}
		if err := interp.assignTo(e.Operand, newF, scope); err != nil {
			return nil, err
		}
		if e.Postfix {
			return oldF, nil
		}
		return newF, nil
	}

	val, err := interp.evalExpr(e.Operand, scope)
	if err != nil {
		return nil, err
	}
	switch opStr {
	case "!":
		return !isTruthy(val), nil
	case "-":
		return -toFloat64Val(val), nil
	case "+":
		return toFloat64Val(val), nil
	case "~":
		return float64(^toInt64Val(val)), nil
	case "void":
		return nil, nil
	}
	return nil, fmt.Errorf("unsupported unary operator: %s", opStr)
}

// assignTo writes val to an lvalue expression (Identifier, BracketExpression, DotExpression)
func (interp *Interpreter) assignTo(target ast.Expression, val Value, scope *Scope) error {
	switch t := target.(type) {
	case *ast.Identifier:
		scope.Set(t.Name.String(), val)
		return nil
	case *ast.BracketExpression:
		objVal, err := interp.evalExpr(t.Left, scope)
		if err != nil {
			return err
		}
		keyVal, err := interp.evalExpr(t.Member, scope)
		if err != nil {
			return err
		}
		return interp.setMember(objVal, toKeyString(keyVal), val)
	case *ast.DotExpression:
		objVal, err := interp.evalExpr(t.Left, scope)
		if err != nil {
			return err
		}
		return interp.setMember(objVal, t.Identifier.Name.String(), val)
	}
	return fmt.Errorf("unsupported assignment target: %T", target)
}

func (interp *Interpreter) evalCall(e *ast.CallExpression, scope *Scope) (Value, error) {
	args := make([]Value, 0, len(e.ArgumentList))
	for _, a := range e.ArgumentList {
		v, err := interp.evalExpr(a, scope)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	var thisVal Value
	var fnVal Value
	var err error

	switch callee := e.Callee.(type) {
	case *ast.BracketExpression:
		objVal, oerr := interp.evalExpr(callee.Left, scope)
		if oerr != nil {
			return nil, oerr
		}
		keyVal, kerr := interp.evalExpr(callee.Member, scope)
		if kerr != nil {
			return nil, kerr
		}
		thisVal = objVal
		fnVal, err = interp.getMember(objVal, toKeyString(keyVal))
		if err != nil {
			return nil, err
		}
	case *ast.DotExpression:
		objVal, oerr := interp.evalExpr(callee.Left, scope)
		if oerr != nil {
			return nil, oerr
		}
		thisVal = objVal
		fnVal, err = interp.getMember(objVal, callee.Identifier.Name.String())
		if err != nil {
			return nil, err
		}
	default:
		fnVal, err = interp.evalExpr(e.Callee, scope)
		if err != nil {
			return nil, err
		}
	}

	return interp.callFunction(fnVal, thisVal, args)
}

func (interp *Interpreter) callFunction(fnVal Value, thisVal Value, args []Value) (Value, error) {
	switch fn := fnVal.(type) {
	case *Function:
		callScope := NewScope(fn.Closure)
		for idx, p := range fn.Params {
			if idx < len(args) {
				callScope.Declare(p, args[idx])
			} else {
				callScope.Declare(p, nil)
			}
		}
		callScope.Declare("this", thisVal)
		callScope.Declare("arguments", append([]Value{}, args...))

		interp.hoist(fn.Body.List, callScope)
		for _, stmt := range fn.Body.List {
			err := interp.execStmt(stmt, callScope)
			if err == nil {
				continue
			}
			if rs, ok := err.(returnSignal); ok {
				return rs.val, nil
			}
			return nil, err
		}
		return nil, nil
	case NativeFunc:
		return fn(args)
	default:
		return nil, fmt.Errorf("attempt to call non-function value: %T", fnVal)
	}
}

func jsTypeOf(v Value) string {
	switch v.(type) {
	case nil:
		return "undefined"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case *Function, NativeFunc:
		return "function"
	default:
		return "object"
	}
}

func nativeStringMethod(s string, method string) (Value, error) {
	switch method {
	case "charAt":
		return NativeFunc(func(args []Value) (Value, error) {
			idx := 0
			if len(args) > 0 {
				idx = int(toFloat64Val(args[0]))
			}
			if idx < 0 || idx >= len(s) {
				return "", nil
			}
			return string(s[idx]), nil
		}), nil
	case "charCodeAt":
		return NativeFunc(func(args []Value) (Value, error) {
			idx := 0
			if len(args) > 0 {
				idx = int(toFloat64Val(args[0]))
			}
			if idx < 0 || idx >= len(s) {
				return math.NaN(), nil
			}
			return float64(s[idx]), nil
		}), nil
	case "slice", "substring":
		return NativeFunc(func(args []Value) (Value, error) {
			start := 0
			end := len(s)
			if len(args) > 0 {
				start = int(toFloat64Val(args[0]))
			}
			if len(args) > 1 {
				end = int(toFloat64Val(args[1]))
			}
			if start < 0 {
				start = len(s) + start
			}
			if end < 0 {
				end = len(s) + end
			}
			if start < 0 {
				start = 0
			}
			if end > len(s) {
				end = len(s)
			}
			if start > end {
				return "", nil
			}
			return s[start:end], nil
		}), nil
	case "indexOf":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return float64(-1), nil
			}
			sub := toStringVal(args[0])
			return float64(strings.Index(s, sub)), nil
		}), nil
	case "split":
		return NativeFunc(func(args []Value) (Value, error) {
			sep := ""
			if len(args) > 0 {
				sep = toStringVal(args[0])
			}
			parts := strings.Split(s, sep)
			out := make([]Value, len(parts))
			for i, p := range parts {
				out[i] = p
			}
			return out, nil
		}), nil
	case "toLowerCase":
		return NativeFunc(func(args []Value) (Value, error) {
			return strings.ToLower(s), nil
		}), nil
	case "toUpperCase":
		return NativeFunc(func(args []Value) (Value, error) {
			return strings.ToUpper(s), nil
		}), nil
	}
	return nil, nil
}
