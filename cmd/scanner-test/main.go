package main

import (
	"log"
	"fmt"
	"io/ioutil"
	"github.com/naucty/akamai-solver-go/pkg/scanner"
)

func main() {
	script, err := ioutil.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	if err != nil {
		log.Fatalf("Failed to read script: %v", err)
	}
	
	s := scanner.NewScanner(string(script))
	dispatchers, err := s.FindDispatchers()
	if err != nil {
		log.Fatalf("Failed to scan: %v", err)
	}
	
	fmt.Printf("Found %d dispatchers:\n", len(dispatchers))
	scanner.PrintDispatchersSummary(dispatchers)
	
	// Check for specific critical dispatchers
	if ty, ok := dispatchers["tY"]; ok {
		fmt.Printf("\ntY dispatcher found:\n")
		fmt.Printf("  Type: %d (3=switch-flattening)\n", ty.Type)
		fmt.Printf("  Cases: %d\n", ty.CaseCount)
		fmt.Printf("  Critical ops: %v\n", ty.CriticalOps)
		fmt.Printf("  Pool refs: %v\n", ty.PoolRefs)
	}
}
