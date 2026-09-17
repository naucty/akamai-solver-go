package vm

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// InterpreterState holds the VM state during execution
type InterpreterState struct {
	Script       string
	Constants    map[string]interface{}
	Tables       map[string][]string
	StringPools  map[string][]string
	LocalVars    map[string]int64 // tP, Jt, rm, etc.
	DecodedChars []string
	PC           int64 // program counter / state variable
}

// NewInterpreter creates a new VM interpreter
func NewInterpreter(scriptContent string) (*InterpreterState, error) {
	interp := &InterpreterState{
		Script:      scriptContent,
		Constants:   make(map[string]interface{}),
		Tables:      make(map[string][]string),
		StringPools: make(map[string][]string),
		LocalVars:   make(map[string]int64),
	}

	// Load precomputed constants from JSON
	constData := []byte(`{
  "indirection_tables": {
    "AX": ["K6","ZC","GI","z9","CA"],
    "hM": ["Nl","gD","vC","gj","YO","fj","wD","kW","x4","JW","X7","KA","Hp","IO","qF","kC","Oj","Tj","dO","MO"]
  },
  "string_pools": {}
}`)
	var consts map[string]interface{}
	if err := json.Unmarshal(constData, &consts); err != nil {
		return nil, fmt.Errorf("failed to parse constants: %w", err)
	}

	// Populate indirection tables
	if tables, ok := consts["indirection_tables"].(map[string]interface{}); ok {
		for name, items := range tables {
			arr := make([]string, 0)
			for _, item := range items.([]interface{}) {
				arr = append(arr, item.(string))
			}
			interp.Tables[name] = arr
		}
	}

	// Initialize constants based on script analysis
	// These values come from constant-folding case Dl
	interp.LocalVars["tP"] = 1   // +![]
	interp.LocalVars["nq"] = 2   // tP+tP
	interp.LocalVars["lV"] = 3   // tP+nq
	interp.LocalVars["ZU"] = 5   // lV+nq
	interp.LocalVars["hP"] = 7   // ZU*tP+nq
	interp.LocalVars["rm"] = 4   // lV+tP
	// Jt, MY, GV must be resolved from context (case KX, lj in Bk)
	// For now, hardcode from earlier resolution:
	interp.LocalVars["Jt"] = 8
	interp.LocalVars["MY"] = 6
	interp.LocalVars["GV"] = 33

	return interp, nil
}

// ResolveIndirection resolves array indexing like hM()[z6]
func (i *InterpreterState) ResolveIndirection(tableName string, index int64) (string, error) {
	table, ok := i.Tables[tableName]
	if !ok {
		return "", fmt.Errorf("table %s not found", tableName)
	}
	if index < 0 || index >= int64(len(table)) {
		return "", fmt.Errorf("index %d out of bounds for table %s (len %d)", index, tableName, len(table))
	}
	return table[index], nil
}

// ExtractPoolFromScript finds encoded string pools in the script
// Pattern: case j8:{var dd=CL[wl]; nI(dd[kq]); qG+=hg; var jh=kq;}
// This locates the pool variable referenced by CL[wl] offset
func (i *InterpreterState) ExtractPoolFromScript() ([]string, error) {
	// Find case j8 block
	re := regexp.MustCompile(`case j8:\{[^}]*\}break;`)
	match := re.FindString(i.Script)
	if match == "" {
		return nil, fmt.Errorf("case j8 not found")
	}

	// For now, return empty pool - real extraction requires more context
	// from the script and CL array reference resolution
	return []string{}, nil
}

// ExecuteDecoder simulates the tY decoder state machine
// Minimal execution: just the XOR loop (case cp)
func (i *InterpreterState) ExecuteDecoder() error {
	// This is where we would step through:
	// case bR: init pool pointer
	// case E: loop through pool
	// case cp: XOR decode each char
	// case kO: return result

	fmt.Println("[VM] Executor not yet implemented")
	return nil
}

// GetDecodedOutput returns the decoded strings in order
func (i *InterpreterState) GetDecodedOutput() []string {
	return i.DecodedChars
}
