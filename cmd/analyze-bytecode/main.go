package main

import (
	"fmt"
	"io/ioutil"
	"sort"
)

type OpcodeFreq struct {
	Byte  byte
	Count int
}

func main() {
	// Load all 6 build bytecodes
	builds := []string{
		"/home/dev/projects/akamai-solver/artifacts/bytecode/00f877a9c1a2909d_0.bin",
		"/home/dev/projects/akamai-solver/artifacts/bytecode/3c13144dcb5b83c6_0.bin",
		"/home/dev/projects/akamai-solver/artifacts/bytecode/35fddfbea9c66083_0.bin",
		"/home/dev/projects/akamai-solver/artifacts/bytecode/5bc3c0a7e1c12345_0.bin", // may not exist
		"/home/dev/projects/akamai-solver/artifacts/bytecode/fda56d71f8e7d1a2_0.bin", // may not exist
		"/home/dev/projects/akamai-solver/artifacts/bytecode/c4050d67a5b8c9d1_0.bin", // may not exist
	}

	globalFreq := make(map[byte]int)

	for _, buildPath := range builds {
		bytecode, err := ioutil.ReadFile(buildPath)
		if err != nil {
			fmt.Printf("Skipped %s (not found)\n", buildPath)
			continue
		}

		// Count byte frequencies
		freq := make(map[byte]int)
		for _, b := range bytecode {
			freq[b]++
		}

		fmt.Printf("\n=== %s (%d bytes) ===\n", buildPath[len(buildPath)-50:], len(bytecode))
		
		// Top 20 frequent opcodes
		type kv struct {
			byte  byte
			count int
		}
		var sorted []kv
		for b, c := range freq {
			sorted = append(sorted, kv{b, c})
			globalFreq[b] += c
		}
		sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

		fmt.Println("Top 20 bytes:")
		for i := 0; i < 20 && i < len(sorted); i++ {
			fmt.Printf("  0x%02x: %4d (%.1f%%)\n", sorted[i].byte, sorted[i].count, 
				100.0*float64(sorted[i].count)/float64(len(bytecode)))
		}
	}

	// Global frequency
	fmt.Println("\n=== GLOBAL TOP 30 BYTES (ACROSS ALL BUILDS) ===")
	type kv struct {
		byte  byte
		count int
	}
	var sorted []kv
	for b, c := range globalFreq {
		sorted = append(sorted, kv{b, c})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })

	for i := 0; i < 30 && i < len(sorted); i++ {
		fmt.Printf("0x%02x: %6d\n", sorted[i].byte, sorted[i].count)
	}

	// Likely opcode patterns
	fmt.Println("\n=== LIKELY OPCODES ===")
	fmt.Println("High frequency (>100 global): likely PUSH/POP/NOP")
	fmt.Println("Medium frequency (10-100): likely binary ops, jumps")
	fmt.Println("Low frequency (<10): likely control flow, rare ops")
}
