package main

import (
	"fmt"
	"os"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

func main() {
	data, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/fresh/script_0.js")
	prog, err := parser.ParseFile(nil, "script_0.js", string(data), 0)
	if err != nil {
		fmt.Println("PARSE ERR:", err)
		os.Exit(1)
	}
	fmt.Printf("Parsed %d top-level statements\n", len(prog.Body))
	for i, stmt := range prog.Body {
		if i > 3 {
			fmt.Println("  ...")
			break
		}
		printStmt(stmt, 2)
	}
}

func printStmt(stmt ast.Statement, depth int) {
	indent := func(d int) string {
		s := ""
		for i := 0; i < d; i++ {
			s += " "
		}
		return s
	}
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		fmt.Printf("%s[%d] ExprStmt: ", indent(depth), depth)
		printExpr(s.Expression, depth+2)
	case *ast.VariableStatement:
		fmt.Printf("%s[%d] VarStmt: %d bindings\n", indent(depth), depth, len(s.List))
	case *ast.FunctionDeclaration:
		fmt.Printf("%s[%d] FuncDecl: %s()\n", indent(depth), depth, s.Function.Name.Name)
	default:
		fmt.Printf("%s[%d] %T\n", indent(depth), depth, s)
	}
}

func printExpr(expr ast.Expression, depth int) {
	indent := func(d int) string {
		s := ""
		for i := 0; i < d; i++ {
			s += " "
		}
		return s
	}
	switch e := expr.(type) {
	case *ast.CallExpression:
		fmt.Printf("CallExpr: callee=%T\n", e.Callee)
		if fn, ok := e.Callee.(*ast.FunctionLiteral); ok {
			fmt.Printf("%s  -> FunctionLiteral(params=%d, body=%d stmts)\n", indent(depth), len(fn.ParameterList.List), len(fn.Body.List))
		}
	case *ast.FunctionLiteral:
		fmt.Printf("FunctionLiteral(params=%d, body=%d stmts)\n", len(e.ParameterList.List), len(e.Body.List))
	case *ast.ObjectLiteral:
		fmt.Printf("ObjectLiteral(%d props)\n", len(e.Value))
	default:
		fmt.Printf("%T\n", e)
	}
}
