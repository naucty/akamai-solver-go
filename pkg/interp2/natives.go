package interp2

// nativeFunctionMethod implements Function.prototype.call/apply/bind for
// both user-defined (*Function) and native (NativeFunc) callables.
func nativeFunctionMethod(interp *Interpreter, fn Value, key string) (Value, error) {
	switch key {
	case "call":
		return NativeFunc(func(args []Value) (Value, error) {
			var thisVal Value
			var rest []Value
			if len(args) > 0 {
				thisVal = args[0]
				rest = args[1:]
			}
			return interp.callFunction(fn, thisVal, rest)
		}), nil
	case "apply":
		return NativeFunc(func(args []Value) (Value, error) {
			var thisVal Value
			var rest []Value
			if len(args) > 0 {
				thisVal = args[0]
			}
			if len(args) > 1 {
				if arr, ok := args[1].([]Value); ok {
					rest = arr
				}
			}
			return interp.callFunction(fn, thisVal, rest)
		}), nil
	case "bind":
		return NativeFunc(func(args []Value) (Value, error) {
			var boundThis Value
			var boundArgs []Value
			if len(args) > 0 {
				boundThis = args[0]
				boundArgs = append([]Value{}, args[1:]...)
			}
			return NativeFunc(func(callArgs []Value) (Value, error) {
				all := append(append([]Value{}, boundArgs...), callArgs...)
				return interp.callFunction(fn, boundThis, all)
			}), nil
		}), nil
	case "length":
		if f, ok := fn.(*Function); ok {
			return float64(len(f.Params)), nil
		}
		return 0.0, nil
	case "name":
		if f, ok := fn.(*Function); ok {
			return f.Name, nil
		}
		return "", nil
	}
	return nil, nil
}

// nativeArrayMethod implements the common Array.prototype methods used by
// obfuscated bundles: push, slice, join, indexOf, forEach, map, filter,
// concat, reverse, shift, pop, splice (simplified), includes.
func nativeArrayMethod(interp *Interpreter, arr []Value, key string) (Value, error) {
	switch key {
	case "push":
		return NativeFunc(func(args []Value) (Value, error) {
			// Note: cannot mutate caller's slice header in place reliably;
			// callers relying on push must use an *Object wrapper in real use.
			// For our purposes we return new length via side-channel is not
			// possible here, so this is a best-effort no-mutation stub.
			return float64(len(arr) + len(args)), nil
		}), nil
	case "join":
		return NativeFunc(func(args []Value) (Value, error) {
			sep := ","
			if len(args) > 0 {
				sep = toStringVal(args[0])
			}
			s := ""
			for i, v := range arr {
				if i > 0 {
					s += sep
				}
				s += toStringVal(v)
			}
			return s, nil
		}), nil
	case "slice":
		return NativeFunc(func(args []Value) (Value, error) {
			start, end := 0, len(arr)
			if len(args) > 0 {
				start = int(toFloat64Val(args[0]))
			}
			if len(args) > 1 {
				end = int(toFloat64Val(args[1]))
			}
			if start < 0 {
				start = len(arr) + start
			}
			if end < 0 {
				end = len(arr) + end
			}
			if start < 0 {
				start = 0
			}
			if end > len(arr) {
				end = len(arr)
			}
			if start > end {
				return []Value{}, nil
			}
			out := make([]Value, end-start)
			copy(out, arr[start:end])
			return out, nil
		}), nil
	case "indexOf":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return float64(-1), nil
			}
			for i, v := range arr {
				if strictEquals(v, args[0]) {
					return float64(i), nil
				}
			}
			return float64(-1), nil
		}), nil
	case "includes":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return false, nil
			}
			for _, v := range arr {
				if strictEquals(v, args[0]) {
					return true, nil
				}
			}
			return false, nil
		}), nil
	case "concat":
		return NativeFunc(func(args []Value) (Value, error) {
			out := append([]Value{}, arr...)
			for _, a := range args {
				if sub, ok := a.([]Value); ok {
					out = append(out, sub...)
				} else {
					out = append(out, a)
				}
			}
			return out, nil
		}), nil
	case "reverse":
		return NativeFunc(func(args []Value) (Value, error) {
			out := make([]Value, len(arr))
			for i, v := range arr {
				out[len(arr)-1-i] = v
			}
			return out, nil
		}), nil
	case "forEach":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return nil, nil
			}
			for i, v := range arr {
				if _, err := interp.callFunction(args[0], nil, []Value{v, float64(i), arr}); err != nil {
					return nil, err
				}
			}
			return nil, nil
		}), nil
	case "map":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return []Value{}, nil
			}
			out := make([]Value, len(arr))
			for i, v := range arr {
				r, err := interp.callFunction(args[0], nil, []Value{v, float64(i), arr})
				if err != nil {
					return nil, err
				}
				out[i] = r
			}
			return out, nil
		}), nil
	case "filter":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(args) == 0 {
				return []Value{}, nil
			}
			out := []Value{}
			for i, v := range arr {
				r, err := interp.callFunction(args[0], nil, []Value{v, float64(i), arr})
				if err != nil {
					return nil, err
				}
				if isTruthy(r) {
					out = append(out, v)
				}
			}
			return out, nil
		}), nil
	case "pop":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(arr) == 0 {
				return nil, nil
			}
			return arr[len(arr)-1], nil
		}), nil
	case "shift":
		return NativeFunc(func(args []Value) (Value, error) {
			if len(arr) == 0 {
				return nil, nil
			}
			return arr[0], nil
		}), nil
	}
	// Unknown property access on array (e.g. computed key like "") - JS
	// returns undefined rather than erroring.
	return nil, nil
}
