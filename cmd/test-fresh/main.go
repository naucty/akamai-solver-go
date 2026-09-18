package main

import (
	"fmt"
	"os"
	"time"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	data, err := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/fresh/script_0.js")
	if err != nil {
		fmt.Println("ERR load:", err)
		os.Exit(1)
	}

	start := time.Now()
	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	_, err = interp.ParseAndRun(string(data))
	elapsed := time.Since(start)
	if err != nil {
		fmt.Printf("ERR exec: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Script executed in %v\n", elapsed)
	if err != nil {
		fmt.Printf("ERR exec: %v\n", err)
		if interp.CallDepth > 0 {
			fmt.Printf("(call depth: %d — likely infinite loop/trampoline)\n", interp.CallDepth)
		}
		os.Exit(1)
	}
}
