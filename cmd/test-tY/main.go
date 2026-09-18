package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	// Load the full har_script.js
	data, err := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	if err != nil {
		fmt.Println("ERR load:", err)
		os.Exit(1)
	}

	// Parse and run the entire script to populate scope
	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	code := string(data)
	_, err = interp.ParseAndRun(code)
	if err != nil {
		fmt.Println("ERR exec:", err)
		os.Exit(1)
	}

	// Find Eh (tY) in the global scope
	globalEh := interp.GlobalScope.Resolve("Eh")
	if globalEh == nil {
		fmt.Println("Eh not in global scope - it's scoped inside the IIFE closure")
		// Try to find it from the scope returned by execStmt for the big IIFE
		// Since we can't directly extract it, we'll need to call through JS

		// Instead, let's call a synthetic function that invokes tY with our test args
		// via the interp's eval engine

		testCode := `
		var result;
		(function(){
			var testState = 0;
			var testArgs = [[1,2,3], "hello"];
			try {
				result = Eh(testState, testArgs);
			} catch(e) {
				result = "ERROR: " + e.message;
			}
		})();
		result;
		`
		_, err := interp.ParseAndRun(testCode)
		if err != nil {
			fmt.Println("ERR test invoke:", err)
			os.Exit(1)
		}
		// Check result in scope
		res := interp.GlobalScope.Resolve("result")
		fmt.Printf("tY result type: %T, value: %v\n", res, res)

		jsonRes, _ := json.MarshalIndent(res, "", "  ")
		fmt.Printf("tY result JSON: %s\n", string(jsonRes))
		return
	}

	fmt.Println("Eh is globally accessible (unexpected, usually private in IIFE)")
	if fn, ok := globalEh.(*interp2.Function); ok {
		fmt.Printf("Eh is a Function: name=%s, params=%v\n", fn.Name, fn.Params)
	}
}
