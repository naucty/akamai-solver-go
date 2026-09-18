package interp2

import (
	"math"
	"strings"
)

// InstallGlobals populates the interpreter's Global scope with the minimal
// set of native builtins needed to execute typical obfuscated browser JS:
// Array, Object, String, Math, window/self/globalThis, console (no-op),
// and JSON (stringify/parse simplified).
func (interp *Interpreter) InstallGlobals() {
	g := interp.Global

	// Array constructor: new Array(n) or Array(a,b,c) or Array.isArray
	arrayProto := &Object{Props: map[string]Value{}}
	arrayCtor := &Object{
		Props: map[string]Value{
			"isArray": NativeFunc(func(args []Value) (Value, error) {
				if len(args) == 0 {
					return false, nil
				}
				_, ok := args[0].([]Value)
				return ok, nil
			}),
			"prototype": arrayProto,
		},
		Call: func(args []Value) (Value, error) {
			if len(args) == 1 {
				if n, ok := args[0].(float64); ok {
					return make([]Value, int(n)), nil
				}
			}
			return append([]Value{}, args...), nil
		},
	}
	g.Declare("Array", arrayCtor)

	// Object constructor + Object.defineProperty/keys/assign/getPrototypeOf
	objectProto := &Object{Props: map[string]Value{}}
	objectCtor := &Object{
		Props: map[string]Value{
			"prototype": objectProto,
			"defineProperty": NativeFunc(func(args []Value) (Value, error) {
				if len(args) < 3 {
					return nil, nil
				}
				target, ok := args[0].(*Object)
				if !ok {
					return args[0], nil // e.g. Array.prototype approximated as Object; ignore silently
				}
				key := toStringVal(args[1])
				descriptor, ok := args[2].(*Object)
				if ok {
					if v, exists := descriptor.Props["value"]; exists {
						target.Props[key] = v
					} else if getter, exists := descriptor.Props["get"]; exists {
						target.Props[key] = getter
					}
				}
				return target, nil
			}),
			"keys": NativeFunc(func(args []Value) (Value, error) {
				if len(args) == 0 {
					return []Value{}, nil
				}
				obj, ok := args[0].(*Object)
				if !ok {
					return []Value{}, nil
				}
				out := make([]Value, 0, len(obj.Props))
				for k := range obj.Props {
					out = append(out, k)
				}
				return out, nil
			}),
			"assign": NativeFunc(func(args []Value) (Value, error) {
				if len(args) == 0 {
					return &Object{Props: map[string]Value{}}, nil
				}
				target, ok := args[0].(*Object)
				if !ok {
					target = &Object{Props: map[string]Value{}}
				}
				for _, src := range args[1:] {
					if so, ok := src.(*Object); ok {
						for k, v := range so.Props {
							target.Props[k] = v
						}
					}
				}
				return target, nil
			}),
			"getPrototypeOf": NativeFunc(func(args []Value) (Value, error) {
				return objectProto, nil
			}),
			"freeze": NativeFunc(func(args []Value) (Value, error) {
				if len(args) > 0 {
					return args[0], nil
				}
				return nil, nil
			}),
		},
		Call: func(args []Value) (Value, error) {
			if len(args) > 0 {
				if o, ok := args[0].(*Object); ok {
					return o, nil
				}
			}
			return &Object{Props: make(map[string]Value)}, nil
		},
	}
	g.Declare("Object", objectCtor)

	// String constructor / String.fromCharCode
	stringObj := &Object{Props: map[string]Value{
		"fromCharCode": NativeFunc(func(args []Value) (Value, error) {
			var sb strings.Builder
			for _, a := range args {
				sb.WriteByte(byte(int(toFloat64Val(a))))
			}
			return sb.String(), nil
		}),
	}}
	g.Declare("String", stringObj)

	// Math object
	mathObj := &Object{Props: map[string]Value{
		"floor": NativeFunc(func(args []Value) (Value, error) {
			return math.Floor(argF(args, 0)), nil
		}),
		"ceil": NativeFunc(func(args []Value) (Value, error) {
			return math.Ceil(argF(args, 0)), nil
		}),
		"round": NativeFunc(func(args []Value) (Value, error) {
			return math.Round(argF(args, 0)), nil
		}),
		"abs": NativeFunc(func(args []Value) (Value, error) {
			return math.Abs(argF(args, 0)), nil
		}),
		"max": NativeFunc(func(args []Value) (Value, error) {
			m := math.Inf(-1)
			for _, a := range args {
				f := toFloat64Val(a)
				if f > m {
					m = f
				}
			}
			return m, nil
		}),
		"min": NativeFunc(func(args []Value) (Value, error) {
			m := math.Inf(1)
			for _, a := range args {
				f := toFloat64Val(a)
				if f < m {
					m = f
				}
			}
			return m, nil
		}),
		"random": NativeFunc(func(args []Value) (Value, error) {
			return 0.5, nil // deterministic for reproducibility
		}),
		"pow": NativeFunc(func(args []Value) (Value, error) {
			return math.Pow(argF(args, 0), argF(args, 1)), nil
		}),
		"PI": math.Pi,
	}}
	g.Declare("Math", mathObj)

	// console (no-op logging)
	noop := NativeFunc(func(args []Value) (Value, error) { return nil, nil })
	consoleObj := &Object{Props: map[string]Value{
		"log": noop, "warn": noop, "error": noop, "debug": noop, "info": noop,
	}}
	g.Declare("console", consoleObj)

	// window / self / globalThis - all point to an object; properties set
	// via SetVar/Declare on Global are separately accessible by name too.
	windowObj := &Object{Props: make(map[string]Value)}
	g.Declare("window", windowObj)
	g.Declare("self", windowObj)
	g.Declare("globalThis", windowObj)

	// Date - minimal, deterministic
	g.Declare("Date", ConstructorFunc(func(args []Value) (Value, error) {
		return &Object{Props: map[string]Value{
			"getTime": NativeFunc(func(a []Value) (Value, error) { return 1694961540000.0, nil }),
		}}, nil
	}))

	// undefined/NaN/Infinity as identifiers some scripts reference directly
	g.Declare("undefined", nil)
	g.Declare("NaN", math.NaN())
	g.Declare("Infinity", math.Inf(1))
}

func argF(args []Value, i int) float64 {
	if i >= len(args) {
		return 0
	}
	return toFloat64Val(args[i])
}
