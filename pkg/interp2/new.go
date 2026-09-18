package interp2

import (
	"fmt"

	"github.com/dop251/goja/ast"
)

// ConstructorFunc creates a new object given constructor args
type ConstructorFunc func(args []Value) (Value, error)

// evalNew handles `new Foo(args)`. We support a small set of native
// constructors (Array, Object, Date, RegExp fallback) plus user functions
// used as constructors (rare in this bytecode but included for completeness).
func (interp *Interpreter) evalNew(e *ast.NewExpression, scope *Scope) (Value, error) {
	args := make([]Value, 0, len(e.ArgumentList))
	for _, a := range e.ArgumentList {
		v, err := interp.evalExpr(a, scope)
		if err != nil {
			return nil, err
		}
		args = append(args, v)
	}

	calleeVal, err := interp.evalExpr(e.Callee, scope)
	if err != nil {
		return nil, err
	}

	switch fn := calleeVal.(type) {
	case ConstructorFunc:
		return fn(args)
	case *Object:
		if fn.Call != nil {
			return fn.Call(args)
		}
		return nil, fmt.Errorf("cannot use 'new' on non-callable object")
	case *Function:
		// naive: create empty object, call fn with `this` bound to it
		obj := &Object{Props: make(map[string]Value)}
		_, err := interp.callFunction(fn, obj, args)
		if err != nil {
			return nil, err
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("cannot use 'new' on non-constructor: %T", calleeVal)
	}
}
