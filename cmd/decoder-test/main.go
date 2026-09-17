package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/naucty/akamai-solver-go/pkg/decoder"
	"github.com/naucty/akamai-solver-go/pkg/scanner"
)

func main() {
	scriptBytes, err := ioutil.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	if err != nil {
		log.Fatalf("Failed to read script: %v", err)
	}
	script := string(scriptBytes)

	s := scanner.NewScanner(script)
	dispatchers, err := s.FindDispatchers()
	if err != nil {
		log.Fatalf("Failed to scan: %v", err)
	}

	tyDisp, ok := dispatchers["tY"]
	if !ok {
		log.Fatalf("tY dispatcher not found")
	}

	fmt.Printf("tY dispatcher: %d cases, %d critical ops\n", tyDisp.CaseCount, len(tyDisp.CriticalOps))
	fmt.Println("Critical op bodies:")
	for name, body := range tyDisp.CriticalOps {
		fmt.Printf("  case %s: %s\n", name, body)
	}

	dc := decoder.NewDecoderChain(script, tyDisp)
	if err := dc.ExecuteCriticalCases(); err != nil {
		fmt.Printf("Execution warning: %v\n", err)
	}

	fmt.Println("\nResolved variables:")
	for name, val := range dc.PoolValues {
		fmt.Printf("  %s = %v\n", name, val)
	}
}
