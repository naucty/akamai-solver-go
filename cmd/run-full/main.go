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
		fmt.Println("read err:", err)
		return
	}

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()
	start := time.Now()
	scope, err := interp.ParseAndRun(string(data))
	elapsed := time.Since(start)
	if err != nil {
		fmt.Println("EXEC ERROR after", elapsed, ":", err)
		if scope != nil {
			// try to see how far we got
		}
		return
	}
	fmt.Println("Full script executed OK in", elapsed)
	if v, ok := scope.Get("tY"); ok {
		fmt.Printf("tY resolved: %T\n", v)
	} else {
		fmt.Println("tY not found in scope")
	}
}
