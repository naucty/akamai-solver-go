package main

import (
	"fmt"
	"os"
	"time"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	data, err := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	if err != nil {
		fmt.Println("ERR load:", err)
		os.Exit(1)
	}

	start := time.Now()
	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	_, err = interp.ParseAndRun(string(data))
	elapsed := time.Since(start)

	fmt.Printf("har_script.js executed in %v\n", elapsed)
	if err != nil {
		fmt.Printf("ERR exec: %v\n", err)
		os.Exit(1)
	}

	// Inspect global scope
	for _, name := range []string{"rnR", "bmak", "get_sensor", "getSensor"} {
		val, ok := interp.Global.Get(name)
		if ok && val != nil {
			fmt.Printf("✓ %s defined: %T\n", name, val)
		}
	}
	fmt.Println("Test PASS")
}
