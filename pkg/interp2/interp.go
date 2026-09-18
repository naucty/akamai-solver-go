// Package interp2 is a JS-subset interpreter walking a real goja/ast tree
// (parsing only via goja/parser - no goja.Runtime, no JS execution engine).
// This replaces the old regex/string-based pkg/interp which could not
// reliably handle nested expressions, closures, or recursive calls.
package interp2

import (
	"fmt"
	"math"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/file"
	"github.com/dop251/goja/parser"
)

// Value is any runtime value: float64, string, bool, nil, []Value, *Object, or *Function
type Value interface{}

// Object is a simple JS-like object (property map). If Call is set, the
// object is also callable (used for constructor-like builtins that also
// carry static properties, e.g. Array.isArray, String.fromCharCode).
type Object struct {
	Props map[string]Value
	Call  func(args []Value) (Value, error)
}

// Function represents a user-defined JS function (closure)
type Function struct {
	Name    string
	Params  []string
	Body    *ast.BlockStatement
	Closure *Scope // lexical scope at definition time
}

// NativeFunc is a Go-implemented builtin exposed to the interpreter
type NativeFunc func(args []Value) (Value, error)

// Scope is a lexical scope (chain of variable environments)
type Scope struct {
	vars   map[string]Value
	parent *Scope
}

func NewScope(parent *Scope) *Scope {
	return &Scope{vars: make(map[string]Value), parent: parent}
}

func (s *Scope) Get(name string) (Value, bool) {
	for sc := s; sc != nil; sc = sc.parent {
		if v, ok := sc.vars[name]; ok {
			return v, true
		}
	}
	return nil, false
}

// Set sets a variable in the scope where it's already declared (closures),
// or in the current scope if not found anywhere (implicit global, sloppy JS)
func (s *Scope) Set(name string, val Value) {
	for sc := s; sc != nil; sc = sc.parent {
		if _, ok := sc.vars[name]; ok {
			sc.vars[name] = val
			return
		}
	}
	s.vars[name] = val
}

// Declare always sets in the CURRENT scope (var/function declarations)
func (s *Scope) Declare(name string, val Value) {
	s.vars[name] = val
}

// Interpreter holds the global scope and native builtins
type Interpreter struct {
	Global *Scope
}

func NewInterpreter() *Interpreter {
	return &Interpreter{Global: NewScope(nil)}
}

// signal types for control flow (break/continue/return) implemented via panic/recover
type breakSignal struct{ label string }
type continueSignal struct{ label string }
type returnSignal struct{ val Value }

// ParseAndRun parses a JS code fragment and executes it in a fresh child scope
// of the interpreter's global scope. Returns the child scope for inspection.
func (interp *Interpreter) ParseAndRun(code string) (*Scope, error) {
	prog, err := parser.ParseFile(nil, "fragment.js", code, 0)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	scope := NewScope(interp.Global)
	if err := interp.hoistAndRun(prog.Body, scope); err != nil {
		return scope, err
	}
	return scope, nil
}

// ParseAndRunDebug executes top-level statements one at a time and reports
// which statement index/type failed, for diagnosing large scripts.
func (interp *Interpreter) ParseAndRunDebug(code string) (*Scope, error) {
	prog, err := parser.ParseFile(nil, "fragment.js", code, 0)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	scope := NewScope(interp.Global)
	interp.hoist(prog.Body, scope)
	for i, stmt := range prog.Body {
		if err := interp.execStmt(stmt, scope); err != nil {
			return scope, fmt.Errorf("stmt[%d] (%T): %w", i, stmt, err)
		}
	}
	return scope, nil
}

// RunInScope executes already-parsed statements in a given scope (for recursive calls)
func (interp *Interpreter) hoistAndRun(stmts []ast.Statement, scope *Scope) error {
	// Hoist var declarations and function declarations first (JS semantics)
	interp.hoist(stmts, scope)

	for _, stmt := range stmts {
		if err := interp.execStmt(stmt, scope); err != nil {
			return err
		}
	}
	return nil
}

func (interp *Interpreter) hoist(stmts []ast.Statement, scope *Scope) {
	for _, stmt := range stmts {
		switch s := stmt.(type) {
		case *ast.VariableStatement:
			for _, binding := range s.List {
				if id, ok := binding.Target.(*ast.Identifier); ok {
					if _, exists := scope.vars[id.Name.String()]; !exists {
						scope.Declare(id.Name.String(), Value(nil))
					}
				}
			}
		case *ast.FunctionDeclaration:
			if s.Function != nil && s.Function.Name != nil {
				fn := interp.makeFunction(s.Function, scope)
				scope.Declare(s.Function.Name.Name.String(), fn)
			}
		case *ast.BlockStatement:
			interp.hoist(s.List, scope)
		case *ast.IfStatement:
			if s.Consequent != nil {
				interp.hoist([]ast.Statement{s.Consequent}, scope)
			}
			if s.Alternate != nil {
				interp.hoist([]ast.Statement{s.Alternate}, scope)
			}
		case *ast.ForStatement:
			if s.Body != nil {
				interp.hoist([]ast.Statement{s.Body}, scope)
			}
		case *ast.WhileStatement:
			if s.Body != nil {
				interp.hoist([]ast.Statement{s.Body}, scope)
			}
		case *ast.DoWhileStatement:
			if s.Body != nil {
				interp.hoist([]ast.Statement{s.Body}, scope)
			}
		}
	}
}

func (interp *Interpreter) makeFunction(fn *ast.FunctionLiteral, closure *Scope) *Function {
	params := make([]string, 0, len(fn.ParameterList.List))
	for _, p := range fn.ParameterList.List {
		if id, ok := p.Target.(*ast.Identifier); ok {
			params = append(params, id.Name.String())
		}
	}
	name := ""
	if fn.Name != nil {
		name = fn.Name.Name.String()
	}
	fnClosure := closure
	if name != "" {
		// Named function expressions can reference themselves by name from
		// inside their own body (e.g. `var YZ = function h9(...){ ... h9(...) }`).
		// Bind the name in a thin scope wrapping the definition closure.
		fnClosure = NewScope(closure)
	}
	f := &Function{Name: name, Params: params, Body: fn.Body, Closure: fnClosure}
	if name != "" {
		fnClosure.Declare(name, f)
	}
	return f
}

// execStmt executes a single statement in the given scope
func (interp *Interpreter) execStmt(stmt ast.Statement, scope *Scope) error {
	switch s := stmt.(type) {
	case nil:
		return nil
	case *ast.VariableStatement:
		for _, binding := range s.List {
			id, ok := binding.Target.(*ast.Identifier)
			if !ok {
				continue
			}
			var val Value
			if binding.Initializer != nil {
				v, err := interp.evalExpr(binding.Initializer, scope)
				if err != nil {
					return err
				}
				val = v
			}
			scope.Declare(id.Name.String(), val)
		}
		return nil

	case *ast.ExpressionStatement:
		_, err := interp.evalExpr(s.Expression, scope)
		return err

	case *ast.BlockStatement:
		child := NewScope(scope) // JS block scoping approx (let/const); var hoisting handled separately
		interp.hoist(s.List, scope) // vars still hoist to function scope in real JS; approximate by hoisting to parent too
		for _, inner := range s.List {
			if err := interp.execStmt(inner, child); err != nil {
				return err
			}
		}
		return nil

	case *ast.IfStatement:
		cond, err := interp.evalExpr(s.Test, scope)
		if err != nil {
			return err
		}
		if isTruthy(cond) {
			return interp.execStmt(s.Consequent, scope)
		} else if s.Alternate != nil {
			return interp.execStmt(s.Alternate, scope)
		}
		return nil

	case *ast.WhileStatement:
		for {
			cond, err := interp.evalExpr(s.Test, scope)
			if err != nil {
				return err
			}
			if !isTruthy(cond) {
				break
			}
			if err := interp.execLoopBody(s.Body, scope); err != nil {
				if _, ok := err.(breakSignal); ok {
					break
				}
				if _, ok := err.(continueSignal); ok {
					continue
				}
				return err
			}
		}
		return nil

	case *ast.DoWhileStatement:
		for {
			if err := interp.execLoopBody(s.Body, scope); err != nil {
				if _, ok := err.(breakSignal); ok {
					break
				}
				if _, ok := err.(continueSignal); ok {
					// fall through to re-check condition
				} else {
					return err
				}
			}
			cond, err := interp.evalExpr(s.Test, scope)
			if err != nil {
				return err
			}
			if !isTruthy(cond) {
				break
			}
		}
		return nil

	case *ast.ForStatement:
		forScope := NewScope(scope)
		if s.Initializer != nil {
			if err := interp.execForInit(s.Initializer, forScope); err != nil {
				return err
			}
		}
		for {
			if s.Test != nil {
				cond, err := interp.evalExpr(s.Test, forScope)
				if err != nil {
					return err
				}
				if !isTruthy(cond) {
					break
				}
			}
			if err := interp.execLoopBody(s.Body, forScope); err != nil {
				if _, ok := err.(breakSignal); ok {
					break
				}
				if _, ok := err.(continueSignal); ok {
					// continue to update
				} else {
					return err
				}
			}
			if s.Update != nil {
				if _, err := interp.evalExpr(s.Update, forScope); err != nil {
					return err
				}
			}
		}
		return nil

	case *ast.ReturnStatement:
		var val Value
		if s.Argument != nil {
			v, err := interp.evalExpr(s.Argument, scope)
			if err != nil {
				return err
			}
			val = v
		}
		return returnSignal{val: val}

	case *ast.BranchStatement:
		if s.Token.String() == "break" {
			label := ""
			if s.Label != nil {
				label = s.Label.Name.String()
			}
			return breakSignal{label: label}
		}
		label := ""
		if s.Label != nil {
			label = s.Label.Name.String()
		}
		return continueSignal{label: label}

	case *ast.SwitchStatement:
		return interp.execSwitch(s, scope)

	case *ast.FunctionDeclaration:
		return nil // already hoisted

	case *ast.EmptyStatement:
		return nil

	case *ast.LabelledStatement:
		err := interp.execStmt(s.Statement, scope)
		if bs, ok := err.(breakSignal); ok && bs.label == s.Label.Name.String() {
			return nil
		}
		return err

	default:
		return fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

func (interp *Interpreter) execLoopBody(body ast.Statement, scope *Scope) error {
	err := interp.execStmt(body, scope)
	if err == nil {
		return nil
	}
	// propagate signals (break/continue/return) up; caller decides
	switch err.(type) {
	case breakSignal, continueSignal, returnSignal:
		return err
	default:
		return err
	}
}

func (interp *Interpreter) execForInit(init ast.ForLoopInitializer, scope *Scope) error {
	switch s := init.(type) {
	case *ast.ForLoopInitializerExpression:
		_, err := interp.evalExpr(s.Expression, scope)
		return err
	case *ast.ForLoopInitializerVarDeclList:
		for _, binding := range s.List {
			id, ok := binding.Target.(*ast.Identifier)
			if !ok {
				continue
			}
			var val Value
			if binding.Initializer != nil {
				v, err := interp.evalExpr(binding.Initializer, scope)
				if err != nil {
					return err
				}
				val = v
			}
			scope.Declare(id.Name.String(), val)
		}
		return nil
	}
	return nil
}

func (interp *Interpreter) execSwitch(s *ast.SwitchStatement, scope *Scope) error {
	disc, err := interp.evalExpr(s.Discriminant, scope)
	if err != nil {
		return err
	}
	matched := false
	var defaultIdx = -1
	for idx, c := range s.Body {
		if !matched {
			if c.Test == nil {
				defaultIdx = idx
				continue
			}
			tv, err := interp.evalExpr(c.Test, scope)
			if err != nil {
				return err
			}
			if looseEquals(disc, tv) {
				matched = true
			}
		}
		if matched {
			for _, cs := range c.Consequent {
				if err := interp.execStmt(cs, scope); err != nil {
					if bs, ok := err.(breakSignal); ok && bs.label == "" {
						return nil
					}
					return err
				}
			}
		}
	}
	if !matched && defaultIdx >= 0 {
		for idx := defaultIdx; idx < len(s.Body); idx++ {
			for _, cs := range s.Body[idx].Consequent {
				if err := interp.execStmt(cs, scope); err != nil {
					if bs, ok := err.(breakSignal); ok && bs.label == "" {
						return nil
					}
					return err
				}
			}
		}
	}
	return nil
}

func (e breakSignal) Error() string    { return "break" }
func (e continueSignal) Error() string { return "continue" }
func (e returnSignal) Error() string   { return "return" }

func isTruthy(v Value) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case float64:
		return x != 0 && !math.IsNaN(x)
	case string:
		return x != ""
	default:
		return true
	}
}

var _ = file.Idx(0) // keep file import used if needed later
