package scanner

import (
	"fmt"
	"regexp"
	"strings"
)

// DispatcherType classifies the obfuscation style
type DispatcherType int

const (
	UnknownDispatcher DispatcherType = iota
	SwitchFlattening  // switch(...){case ...: ... break;}
	BytecodeArray     // array of opcodes with dispatch loop
	RecursiveCall     // E0f(op, [...])
)

// Dispatcher represents an identified dispatcher function
type Dispatcher struct {
	Name         string
	Type         DispatcherType
	Position     int
	CaseCount    int
	CriticalOps  map[string]string // opname -> code snippet
	PoolRefs     map[string]bool    // referenced pool/variable names
}

// Scanner analyzes a script for dispatchers
type Scanner struct {
	Script string
}

// NewScanner creates a new script scanner
func NewScanner(script string) *Scanner {
	return &Scanner{Script: script}
}

// FindDispatchers identifies all dispatcher functions
func (s *Scanner) FindDispatchers() (map[string]*Dispatcher, error) {
	dispatchers := make(map[string]*Dispatcher)
	
	// Find function definitions
	funcRegex := regexp.MustCompile(`function\s+(\w+)\s*\([^)]*\)\s*\{`)
	matches := funcRegex.FindAllStringSubmatchIndex(s.Script, -1)
	
	for _, match := range matches {
		start := match[0]
		nameStart, nameEnd := match[2], match[3]
		
		funcName := s.Script[nameStart:nameEnd]
		
		// Extract function body
		bodyStart := strings.Index(s.Script[start:], "{") + start
		bodyEnd := s.findMatchingBrace(bodyStart)
		
		if bodyEnd == -1 {
			continue
		}
		
		body := s.Script[bodyStart+1 : bodyEnd]
		
		// Detect dispatcher type
		dispType := s.classifyDispatcher(body)
		if dispType == UnknownDispatcher {
			continue
		}
		
		// Analyze the dispatcher
		disp := &Dispatcher{
			Name:        funcName,
			Type:        dispType,
			Position:    start,
			CriticalOps: make(map[string]string),
			PoolRefs:    make(map[string]bool),
		}
		
		// Extract critical operations (cases)
		s.extractCriticalOps(disp, body)
		
		// Extract pool references
		s.extractPoolReferences(disp, body)
		
		dispatchers[funcName] = disp
	}
	
	return dispatchers, nil
}

// classifyDispatcher detects the obfuscation type
func (s *Scanner) classifyDispatcher(body string) DispatcherType {
	// Check for switch-flattening pattern
	if strings.Contains(body, "switch") && strings.Contains(body, "case") && strings.Contains(body, "break;") {
		return SwitchFlattening
	}
	
	// Check for bytecode-array pattern
	if strings.Contains(body, "while") && (strings.Contains(body, "[") || strings.Contains(body, "array")) {
		// Could be bytecode dispatch
		return BytecodeArray
	}
	
	// Check for recursive call pattern
	if regexp.MustCompile(`\w+\s*\(\s*\w+\s*,`).MatchString(body) {
		return RecursiveCall
	}
	
	return UnknownDispatcher
}

// extractCriticalOps extracts case statements for critical operations
func (s *Scanner) extractCriticalOps(disp *Dispatcher, body string) {
	// Find all case statements
	caseRegex := regexp.MustCompile(`case\s+(\w+)\s*:\s*\{([^}]*)\}`)
	matches := caseRegex.FindAllStringSubmatch(body, -1)
	
	disp.CaseCount = len(matches)
	
	// Extract specific critical cases (E, cp, bR, kO, j8, etc.)
	criticalNames := map[string]bool{
		"E": true, "cp": true, "bR": true, "kO": true, "j8": true,
		"vg": true, "ZX": true, "HC": true, "Dl": true, "jW": true,
	}
	
	for _, match := range matches {
		if len(match) >= 3 {
			caseName := match[1]
			caseBody := match[2]
			
			if criticalNames[caseName] {
				disp.CriticalOps[caseName] = caseBody
			}
		}
	}
}

// extractPoolReferences finds variable/pool references in the dispatcher
func (s *Scanner) extractPoolReferences(disp *Dispatcher, body string) {
	// Look for patterns like: var pV = ..., var KI = ..., Td.HQ = ...
	poolPatterns := []string{
		`var\s+(\w{2})\s*=`,  // var XX =
		`(\w{2})\s*=\s*\w+\[`, // XX = [...][...]
	}
	
	varRegex := regexp.MustCompile(strings.Join(poolPatterns, "|"))
	matches := varRegex.FindAllStringSubmatch(body, -1)
	
	for _, match := range matches {
		for i := 1; i < len(match); i++ {
			if match[i] != "" {
				disp.PoolRefs[match[i]] = true
			}
		}
	}
}

// findMatchingBrace finds the closing brace for an opening brace
func (s *Scanner) findMatchingBrace(openPos int) int {
	if openPos < 0 || openPos >= len(s.Script) || s.Script[openPos] != '{' {
		return -1
	}
	
	depth := 1
	for i := openPos + 1; i < len(s.Script); i++ {
		switch s.Script[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	
	return -1
}

// PrintDispatchersSummary prints a summary of found dispatchers
func PrintDispatchersSummary(dispatchers map[string]*Dispatcher) {
	for name, disp := range dispatchers {
		typeStr := ""
		switch disp.Type {
		case SwitchFlattening:
			typeStr = "Switch-Flattening"
		case BytecodeArray:
			typeStr = "Bytecode-Array"
		case RecursiveCall:
			typeStr = "Recursive-Call"
		default:
			typeStr = "Unknown"
		}
		
		fmt.Printf("%s: %s (%d cases)\n", name, typeStr, disp.CaseCount)
		if len(disp.CriticalOps) > 0 {
			fmt.Printf("  critical ops: %v\n", len(disp.CriticalOps))
		}
		if len(disp.PoolRefs) > 0 {
			fmt.Printf("  pool refs: %v\n", disp.PoolRefs)
		}
	}
}
