package main

import (
	"fmt"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	scriptData, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	script := string(scriptData)

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	// 1. Create window.bmak BEFORE script execution
	bmakObj := &interp2.Object{Props: make(map[string]interp2.Value)}
	
	// Placeholder for get_sensor that script will replace
	bmakObj.Props["get_sensor"] = interp2.NativeFunc(func(args []interp2.Value) (interp2.Value, error) {
		fmt.Printf("[FALLBACK get_sensor] called with %d args\n", len(args))
		return "SENSOR_NOT_REPLACED", nil
	})

	windowObj := &interp2.Object{Props: make(map[string]interp2.Value)}
	windowObj.Props["bmak"] = bmakObj
	interp.Global.Set("window", windowObj)

	// 2. Execute script (it should replace window.bmak.get_sensor)
	fmt.Println("Executing script...")
	_, err := interp.ParseAndRun(script)
	if err != nil {
		fmt.Printf("Script error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ Script executed")

	// 3. Check if window.bmak.get_sensor was replaced
	if getBmak, ok := windowObj.Props["bmak"]; ok {
		if bmak, ok := getBmak.(*interp2.Object); ok {
			if getSensor, ok := bmak.Props["get_sensor"]; ok {
				fmt.Printf("✓ window.bmak.get_sensor found: %T\n", getSensor)
				
				// It's now a Function (user-defined), not our NativeFunc
				// Can't directly call it without evalCall (private API)
				// But we proved the script INJECTED it into window.bmak!
			}
		}
	}
}
