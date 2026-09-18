package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	scriptData, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	script := string(scriptData)

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	// Will be captured when script runs
	var capturedEh interp2.Value
	var sensorOutput string

	// Capture Eh, HC, hs from inside IIFE
	interp.Global.Set("__capture__", interp2.NativeFunc(func(args []interp2.Value) (interp2.Value, error) {
		fmt.Printf("__capture__ called with %d args\n", len(args))
		if len(args) > 0 {
			capturedEh = args[0]
			fmt.Printf("  arg[0] (Eh): %T\n", args[0])
		}
		return nil, nil
	}))

	// Capture sensor result
	interp.Global.Set("__captureSensor__", interp2.NativeFunc(func(args []interp2.Value) (interp2.Value, error) {
		if len(args) > 0 {
			sensorOutput = fmt.Sprintf("%v", args[0])
			fmt.Printf("[SENSOR RESULT] %s\n", sensorOutput)
		}
		return nil, nil
	}))

	// Inject capture calls into script
	injectPoint := strings.LastIndex(script, "HB;")
	if injectPoint < 0 {
		fmt.Println("ERR: injection point not found")
		os.Exit(1)
	}

	injectCode := `
console.log("DEBUG: injection code running");
console.log("typeof Eh = " + (typeof Eh));
console.log("typeof HC = " + (typeof HC));
try {
  __capture__(typeof Eh !== 'undefined' ? Eh : null);
  console.log("DEBUG: __capture__ called");
} catch(e) {
  console.log("DEBUG: error calling __capture__: " + e);
}
`
	modScript := script[:injectPoint] + injectCode + script[injectPoint:]

	fmt.Println("Executing script with injected capture...")
	_, err := interp.ParseAndRun(modScript)
	if err != nil {
		fmt.Printf("Script error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== Results ===")
	fmt.Printf("Captured Eh: %T\n", capturedEh)
	fmt.Printf("Sensor output: %s\n", sensorOutput)

	if capturedEh == nil {
		fmt.Println("FAIL: Could not capture Eh")
		os.Exit(1)
	}
}
