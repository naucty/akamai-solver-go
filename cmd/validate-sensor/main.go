package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	// Load captured reference
	refData, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/fresh/sensors.json")
	var sensors []map[string]interface{}
	json.Unmarshal(refData, &sensors)

	refPayload := sensors[0]["sensor"].(string)
	parts := split(refPayload, ";")
	refB64 := parts[5]

	refBytes, _ := base64.StdEncoding.DecodeString(refB64)
	fmt.Printf("Reference (from camoufox):\n")
	fmt.Printf("  B64: %s\n", refB64)
	fmt.Printf("  Bytes (hex): %x\n", refBytes)
	fmt.Printf("  SHA256: %x\n", sha256.Sum256(refBytes))

	// Now execute har_script.js and try to reproduce same bytes
	scriptData, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	_, err := interp.ParseAndRun(string(scriptData))
	if err != nil {
		fmt.Printf("Script exec error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✓ Script executed successfully")
	fmt.Println("Now: need to call tY(qG, CL) to generate sensor bytes")
	fmt.Println("Next: reverse-engineer qG, CL from captured payloads")
}

func split(s, sep string) []string {
	var parts []string
	curr := ""
	for _, c := range s {
		if string(c) == sep {
			parts = append(parts, curr)
			curr = ""
		} else {
			curr += string(c)
		}
	}
	parts = append(parts, curr)
	return parts
}
