package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	// Load script
	scriptData, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	script := string(scriptData)

	// Strategy: Inject export wrapper that runs INSIDE the IIFE
	// before it closes, to capture Eh and hs
	
	// Find the closing of main IIFE
	// Pattern: }());
	iieEnd := strings.LastIndex(script, "HB;")
	if iieEnd < 0 {
		fmt.Println("ERR: Could not find script end")
		os.Exit(1)
	}

	exportCode := `
// === EXPORT WRAPPER (injected) ===
try {
  if (typeof Eh !== 'undefined') {
    window._SENSOR_EXPORT = {};
    window._SENSOR_EXPORT.Eh = Eh;
    window._SENSOR_EXPORT.tY = Eh;
    window._SENSOR_EXPORT.HC = HC;
    window._SENSOR_EXPORT.hs = hs;
    window._SENSOR_EXPORT.bmak = typeof bmak !== 'undefined' ? bmak : null;
  } else {
    window._SENSOR_EXPORT_ERROR = "Eh not defined at injection point";
  }
} catch(e) {
  window._SENSOR_EXPORT_ERROR = String(e);
}
// === END EXPORT ===
`

	modifiedScript := script[:iieEnd] + "\n" + exportCode + "\n" + script[iieEnd:]

	// Execute in interpreter
	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	// Inject mock globals
	interp.Global.Set("window", &interp2.Object{Props: make(map[string]interp2.Value)})

	fmt.Println("Executing script...")
	_, err := interp.ParseAndRun(modifiedScript)
	if err != nil {
		fmt.Printf("Script error: %v\n", err)
		os.Exit(1)
	}

	// Try to access exports
	windowVal, ok := interp.Global.Get("window")
	if !ok || windowVal == nil {
		fmt.Println("ERR: window not found")
		os.Exit(1)
	}

	windowObj, ok := windowVal.(*interp2.Object)
	if !ok {
		fmt.Println("ERR: window is not an Object")
		os.Exit(1)
	}

	exportObj, ok := windowObj.Props["_SENSOR_EXPORT"]
	if !ok || exportObj == nil {
		// Check for error
		if errVal, hasErr := windowObj.Props["_SENSOR_EXPORT_ERROR"]; hasErr {
			fmt.Printf("ERR from injected code: %v\n", errVal)
		} else {
			fmt.Println("ERR: _SENSOR_EXPORT not found in window")
			fmt.Printf("window has %d props\n", len(windowObj.Props))
			for k := range windowObj.Props {
				fmt.Printf("  - %s\n", k)
			}
		}
		os.Exit(1)
	}

	exportObjTyped, ok := exportObj.(*interp2.Object)
	if !ok {
		fmt.Println("ERR: _SENSOR_EXPORT is not an Object")
		os.Exit(1)
	}

	// Check what we got
	fmt.Println("✓ Exports captured:")
	for key, val := range exportObjTyped.Props {
		fmt.Printf("  %s: %T\n", key, val)
	}

	// Now try to call hs with test params
	hsVal, ok := exportObjTyped.Props["hs"]
	if !ok || hsVal == nil {
		fmt.Println("ERR: hs not in exports")
		os.Exit(1)
	}

	fmt.Printf("\nhs type: %T\n", hsVal)

	// TODO: Call hs with params
	// This requires callFunction API in interp2
}
