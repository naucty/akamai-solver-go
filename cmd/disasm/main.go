package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/naucty/akamai-solver-go/pkg/vm"
)

func main() {
	dir := "/home/dev/projects/akamai-solver/artifacts/bytecode"
	files, _ := filepath.Glob(filepath.Join(dir, "*.bin"))
	sort.Strings(files)

	for _, fpath := range files {
		code, _ := os.ReadFile(fpath)
		finder := &vm.StringFinder{Code: code}
		ops := finder.FindStringOpcodes(8)

		opsMap := make(map[byte]bool)
		for _, o := range ops {
			opsMap[o] = true
		}

		instrs := finder.Disassemble(opsMap)
		strs := vm.ExtractStrings(instrs)
		norm := vm.Normalize(strs)

		fmt.Printf("=== %s (%d bytes) ===\n", filepath.Base(fpath), len(code))
		fmt.Printf("String opcodes: ")
		for _, o := range ops {
			fmt.Printf("%02x ", o)
		}
		fmt.Printf("\nInstructions: %d, Strings: %d\n", len(instrs), len(strs))
		fmt.Printf("Normalized (first 25):\n")
		for i, n := range norm {
			if i > 25 {
				fmt.Printf("  ... (%d more)\n", len(norm)-25)
				break
			}
			fmt.Printf("  [%2d] %s\n", i, n)
		}
		fmt.Println()
	}
}
