package main

import (
	"fmt"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	// Load har_script with export wrapper
	data, err := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	if err != nil {
		fmt.Println("ERR load:", err)
		os.Exit(1)
	}

	// Append export at end of script
	exportCode := `
// Export for Go harness
var _TY_EXPORT = typeof tY !== 'undefined' ? tY : (typeof rnR !== 'undefined' ? rnR : null);
var _EXPORT = {
  tY: _TY_EXPORT,
  rnR: typeof rnR !== 'undefined' ? rnR : null,
  bmak: typeof bmak !== 'undefined' ? bmak : null
};
_EXPORT;
`
	fullScript := string(data) + "\n" + exportCode

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	result, err := interp.ParseAndRun(fullScript)
	if err != nil {
		fmt.Printf("Execution error: %v\n", err)
		os.Exit(1)
	}

	_ = result
	fmt.Printf("ParseAndRun result scope type: %T\n", result)

	// Check what the script RETURNS as a whole value
	retVal := interp.LastExprValue
	fmt.Printf("Script last expr value: %T %v\n", retVal, retVal)

	// Check exports
	exportObj, ok := result.Get("_EXPORT")
		fmt.Printf("✓ _EXPORT defined: %T\n", exportObj)
		if obj, isObj := exportObj.(*interp2.Object); isObj {
			fmt.Printf("  Properties:\n")
			for key, val := range obj.Props {
				fmt.Printf("    %s: %T\n", key, val)
			}
		}
	}

	for _, name := range []string{"tY", "rnR", "bmak"} {
		val, ok := result.Get(name)
		if ok && val != nil {
			fmt.Printf("✓ %s in global scope: %T\n", name, val)
		}
	}
}
