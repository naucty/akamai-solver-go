package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	// Load har_script
	data, err := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	if err != nil {
		fmt.Println("ERR load script:", err)
		os.Exit(1)
	}

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	// Install tY hook BEFORE execution
	var capturedCalls []map[string]interface{}
	
	// Create a NativeFunc that wraps tY calls
	tYHook := interp2.NativeFunc(func(args []interp2.Value) (interp2.Value, error) {
		call := map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"args_len": len(args),
		}
		if len(args) > 0 {
			call["arg0"] = fmt.Sprintf("%v", args[0])
		}
		if len(args) > 1 {
			call["arg1"] = fmt.Sprintf("%v", args[1])
		}
		capturedCalls = append(capturedCalls, call)
		fmt.Printf("tY called with %d args\n", len(args))
		return nil, nil
	})

	// Set tY in global scope before parsing
	interp.Global.Set("tY", tYHook)

	// Parse and run (this will initialize the script, may trigger tY calls)
	_, err = interp.ParseAndRun(string(data))
	if err != nil {
		fmt.Printf("Script execution: %v\n", err)
	}

	// Now tY is already defined, we need to call it
	// Try to find tY in global scope and call it with test args
	tYVal, ok := interp.Global.Get("tY")
	if !ok {
		fmt.Println("ERR: tY not found in global scope")
		os.Exit(1)
	}

	// Try calling tY with common args
	testCases := []struct {
		arg0 string
		arg1 string
	}{
		{"", ""},
		{"test", "test"},
	}

	for _, tc := range testCases {
		fmt.Printf("\nCalling tY(%q, %q)\n", tc.arg0, tc.arg1)
		_, _ = interp.CallFunction(tYVal, []interp2.Value{
			interp2.NewString(tc.arg0),
			interp2.NewString(tc.arg1),
		})

	// Output captured calls
	if len(capturedCalls) > 0 {
		b, _ := json.MarshalIndent(capturedCalls, "", "  ")
		fmt.Printf("\nCaptured calls:\n%s\n", b)
	} else {
		fmt.Println("\nNo tY calls captured during execution")
	}
}
