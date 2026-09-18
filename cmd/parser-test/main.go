package main

import (
	"fmt"

	"github.com/dop251/goja/ast"
	"github.com/dop251/goja/parser"
)

func main() {
	code := `
	qG -= 5;
	while (vc < pV.length) {
		var gG = pV[vc];
		var WI = keystream[NY++];
		Pd += String.fromCharCode(gG ^ WI);
		vc++;
	}
	`

	prog, err := parser.ParseFile(nil, "test.js", code, 0)
	if err != nil {
		fmt.Println("PARSE ERROR:", err)
		return
	}

	fmt.Printf("Parsed %d statements\n", len(prog.Body))
	for idx, stmt := range prog.Body {
		fmt.Printf("[%d] %T\n", idx, stmt)
	}

	// Try to walk into the while statement
	for _, stmt := range prog.Body {
		if ws, ok := stmt.(*ast.WhileStatement); ok {
			fmt.Printf("WhileStatement found: test=%T, body=%T\n", ws.Test, ws.Body)
		}
	}
}
