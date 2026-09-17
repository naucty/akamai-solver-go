package decoder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/naucty/akamai-solver-go/pkg/interp"
	"github.com/naucty/akamai-solver-go/pkg/scanner"
)

// DecoderChain orchestrates the interpretation of decoder cases
type DecoderChain struct {
	Script      string
	Interpreter *interp.Interpreter
	Dispatcher  *scanner.Dispatcher
	PoolValues  map[string]interface{} // resolved pool names -> values
}

// NewDecoderChain creates a new decoder chain for a specific dispatcher
func NewDecoderChain(script string, disp *scanner.Dispatcher) *DecoderChain {
	return &DecoderChain{
		Script:      script,
		Interpreter: interp.NewInterpreter(),
		Dispatcher:  disp,
		PoolValues:  make(map[string]interface{}),
	}
}

// ExecuteCriticalCases runs the critical cases in sequence to resolve pools
func (dc *DecoderChain) ExecuteCriticalCases() error {
	if dc.Dispatcher == nil || len(dc.Dispatcher.CriticalOps) == 0 {
		return fmt.Errorf("no critical ops in dispatcher")
	}

	// Order matters: bR (init), E (decode loop), cp (more decoding), kO (return)
	order := []string{"bR", "j8", "E", "cp", "kO"}

	for _, caseName := range order {
		caseBody, ok := dc.Dispatcher.CriticalOps[caseName]
		if !ok {
			continue // This case not in this dispatcher
		}

		fmt.Printf("[DecoderChain] Executing case %s\n", caseName)

		if err := dc.Interpreter.ExecuteCase(caseBody); err != nil {
			fmt.Printf("[DecoderChain] Warning: case %s failed: %v (continuing)\n", caseName, err)
			// Don't fail entirely, some cases might not apply
		}

		// After bR, try to capture pool references
		if caseName == "bR" {
			dc.capturePoolReferences()
		}
	}

	return nil
}

// capturePoolReferences extracts resolved pool variable values
func (dc *DecoderChain) capturePoolReferences() {
	// Look for variables that match pool reference names
	for poolName := range dc.Dispatcher.PoolRefs {
		if val, ok := dc.Interpreter.GetVar(poolName); ok {
			dc.PoolValues[poolName] = val
			fmt.Printf("[DecoderChain] Captured pool %s = %v\n", poolName, val)
		}
	}
}

// ResolvePoolVariable finds the actual value of a pool variable
// by tracing its definition in the script
func (dc *DecoderChain) ResolvePoolVariable(varName string) (interface{}, error) {
	// First check if it's already in the interpreter
	if val, ok := dc.Interpreter.GetVar(varName); ok {
		return val, nil
	}

	// Try to find its definition in the script
	// Pattern: var varName = ... or varName = ...
	pattern := regexp.MustCompile(fmt.Sprintf(`\b%s\s*=\s*([^;,]+)`, regexp.QuoteMeta(varName)))
	matches := pattern.FindStringSubmatch(dc.Script)

	if len(matches) < 2 {
		return nil, fmt.Errorf("could not resolve %s", varName)
	}

	exprStr := strings.TrimSpace(matches[1])
	fmt.Printf("[DecoderChain] Resolving %s = %s\n", varName, exprStr)

	return nil, nil // Stub - would need full context to eval this
}

// ExtractDecodedStrings would implement the actual XOR loop once pools are resolved
func (dc *DecoderChain) ExtractDecodedStrings() ([]string, error) {
	// TODO: Implement XOR decoding loop
	// Use pV (pool), Td.HQ (keystream), XOR operation
	// Return decoded strings in order

	return []string{}, nil
}

// GetResolvedVar returns a variable from the interpreter
func (dc *DecoderChain) GetResolvedVar(name string) (interface{}, bool) {
	return dc.Interpreter.GetVar(name)
}
