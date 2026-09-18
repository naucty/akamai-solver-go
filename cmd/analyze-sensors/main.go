package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func main() {
	data, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/fresh/sensors.json")
	var sensors []map[string]interface{}
	json.Unmarshal(data, &sensors)

	fmt.Println("=== Sensor Payload Analysis ===\n")

	for i, s := range sensors {
		payload := s["sensor"].(string)
		parts := strings.Split(payload, ";")

		fmt.Printf("Payload %d:\n", i)
		fmt.Printf("  Version: %s\n", parts[0])
		fmt.Printf("  Reserved: %s\n", parts[1])
		fmt.Printf("  Unknown1: %s\n", parts[2])
		fmt.Printf("  Unknown2: %s\n", parts[3])
		fmt.Printf("  Seed: %s\n", parts[4])

		// Decode base64
		if len(parts) > 5 {
			b64 := parts[5]
			fmt.Printf("  B64 input: %s (len=%d)\n", b64, len(b64))

			decoded, err := base64.StdEncoding.DecodeString(b64)
			if err != nil {
				fmt.Printf("  ERR decode: %v\n", err)
			} else {
				fmt.Printf("  Decoded bytes (len=%d): %v\n", len(decoded), decoded)
				fmt.Printf("  Hex: %x\n", decoded)
			}
		}

		// Metrics
		if len(parts) > 6 {
			fmt.Printf("  Metrics: %s\n", parts[6])
		}

		// Obfuscated data length
		if len(parts) > 7 {
			fmt.Printf("  Obfuscated len: %d bytes\n", len(parts[7]))
		}
		fmt.Println()
	}
}
