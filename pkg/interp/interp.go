package interp

import (
	"fmt"
	"regexp"
	"strings"
)

// Value represents a runtime value in the interpreter
type Value interface{}

// Interpreter holds execution state
type Interpreter struct {
	Vars   map[string]Value // local variables
	Global map[string]Value // global/closure scope
	Code   string           // source code fragment
}

// NewInterpreter creates a new JS-subset interpreter
func NewInterpreter() *Interpreter {
	return &Interpreter{
		Vars:   make(map[string]Value),
		Global: make(map[string]Value),
	}
}

// GetVar retrieves a variable value
func (i *Interpreter) GetVar(name string) (Value, bool) {
	if val, ok := i.Vars[name]; ok {
		return val, true
	}
	if val, ok := i.Global[name]; ok {
		return val, true
	}
	return nil, false
}

// SetVar sets a variable value
func (i *Interpreter) SetVar(name string, value Value) {
	i.Vars[name] = value
}

// ExecuteCase executes a single switch case block (string of JS code)
// Example: `var tP=+ ! ![]; var nq=tP+tP;` (from case Dl)
func (i *Interpreter) ExecuteCase(codeBlock string) error {
	i.Code = codeBlock
	return i.execute()
}

// execute runs the code block
func (i *Interpreter) execute() error {
	// Split by semicolons to get statements
	statements := strings.Split(i.Code, ";")
	
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		
		// Check if this is a loop (must be before executeStatement)
		if strings.HasPrefix(stmt, "while") || strings.HasPrefix(stmt, "do") || strings.HasPrefix(stmt, "for") {
			if err := i.handleLoop(stmt); err != nil {
				return fmt.Errorf("loop failed: %w", err)
			}
			continue
		}
		
		// Try to parse and execute as regular statement
		if err := i.executeStatement(stmt); err != nil {
			return fmt.Errorf("failed to execute '%s': %w", stmt, err)
		}
	}
	
	return nil
}

// executeStatement handles a single statement
func (i *Interpreter) executeStatement(stmt string) error {
	stmt = strings.TrimSpace(stmt)
	
	// Handle variable declaration + assignment: var x = expr;
	if strings.HasPrefix(stmt, "var ") {
		return i.handleVarDecl(stmt)
	}
	
	// Handle simple assignment: x = expr;
	if strings.Contains(stmt, "=") && !strings.Contains(stmt, "==") && !strings.Contains(stmt, "!=") {
		return i.handleAssignment(stmt)
	}
	
	// Handle while/do-while (simplified for now)
	if strings.Contains(stmt, "while") || strings.Contains(stmt, "do") {
		return i.handleLoop(stmt)
	}
	
	// Handle switch (simplified)
	if strings.Contains(stmt, "switch") {
		return i.handleSwitchStmt(stmt)
	}
	
	return nil // ignore unknown statements for now
}

// handleVarDecl parses `var name = expr` and stores the value
func (i *Interpreter) handleVarDecl(stmt string) error {
	// Extract "var name = expr"
	re := regexp.MustCompile(`var\s+(\w+)\s*=\s*(.+)$`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 3 {
		return fmt.Errorf("invalid var declaration: %s", stmt)
	}
	
	name := matches[1]
	exprStr := strings.TrimSpace(matches[2])
	
	val, err := i.evalExpr(exprStr)
	if err != nil {
		return err
	}
	
	i.Vars[name] = val
	return nil
}

// handleAssignment parses `name = expr` and updates the value
func (i *Interpreter) handleAssignment(stmt string) error {
	parts := strings.Split(stmt, "=")
	if len(parts) < 2 {
		return fmt.Errorf("invalid assignment: %s", stmt)
	}
	
	name := strings.TrimSpace(parts[0])
	exprStr := strings.TrimSpace(strings.Join(parts[1:], "="))
	
	val, err := i.evalExpr(exprStr)
	if err != nil {
		return err
	}
	
	i.Vars[name] = val
	return nil
}


// handleSwitch handles switch statements (stub for now)
func (i *Interpreter) handleSwitchStmt(stmt string) error {
	// TODO: implement switch execution
	return nil
}

// evalExpr evaluates a simple expression and returns its value
// Handles: literals, variable refs, arithmetic (+, -, *, /, %), JSFuck (!, +[], etc.)
func (i *Interpreter) evalExpr(exprStr string) (Value, error) {
	exprStr = strings.TrimSpace(exprStr)
	
	// Try to evaluate JSFuck-style literals first
	if val, ok := i.evalJSFuck(exprStr); ok {
		return val, nil
	}
	
	// Try as numeric literal
	if val, err := parseNumber(exprStr); err == nil {
		return val, nil
	}
	
	// Try as variable reference
	if val, exists := i.Vars[exprStr]; exists {
		return val, nil
	}
	if val, exists := i.Global[exprStr]; exists {
		return val, nil
	}
	
	// Try as arithmetic expression
	return i.evalArithmetic(exprStr)
}

// evalJSFuck handles JSFuck-style obfuscation
// Examples: ![] = false, +[] = 0, +!![] = 1, !+[] = false
func (i *Interpreter) evalJSFuck(expr string) (Value, bool) {
	expr = strings.TrimSpace(expr)
	
	// +!![] = +(!(![]) = +(true) = 1
	if expr == "+!![]" || expr == "+! ![]" || expr == "+ ! ![]" {
		return int64(1), true
	}
	
	// ![] = false
	if expr == "![]" {
		return false, true
	}
	
	// +[] = 0
	if expr == "+[]" {
		return int64(0), true
	}
	
	// !+[] = false
	if expr == "!+[]" {
		return false, true
	}
	
	return nil, false
}

// evalArithmetic evaluates expressions like "a+b", "a-b*c", etc.
func (i *Interpreter) evalArithmetic(exprStr string) (Value, error) {
	// Recursively substitute variables and evaluate
	expr := exprStr
	
	// Replace known variables with their values (simple substitution)
	for {
		changed := false
		for name, val := range i.Vars {
			if idx := strings.Index(expr, name); idx >= 0 {
				// Check it's a whole word (not part of another identifier)
				before := idx == 0 || !isIdentifierChar(rune(expr[idx-1]))
				after := idx+len(name) >= len(expr) || !isIdentifierChar(rune(expr[idx+len(name)]))
				
				if before && after {
					expr = expr[:idx] + fmt.Sprintf("(%v)", val) + expr[idx+len(name):]
					changed = true
					break
				}
			}
		}
		if !changed {
			break
		}
	}
	
	// Try to evaluate the substituted expression
	result := evaluateArithmeticExpr(expr)
	if result == nil {
		return nil, fmt.Errorf("unable to evaluate: %s", exprStr)
	}
	return result, nil
}

// isIdentifierChar checks if a rune is part of an identifier
func isIdentifierChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '$'
}

// evaluateArithmeticExpr evaluates a fully-substituted arithmetic expression
// This is a simplified eval; in production you'd want a proper parser

// parseNumber tries to parse a numeric literal
func parseNumber(s string) (int64, error) {
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// mustParseInt parses an integer, panicking on failure (for internal use)
func mustParseInt(s string) int64 {
	s = strings.TrimSpace(strings.Trim(s, "()"))
	val, _ := parseNumber(s)
	return val
}

// GetVar retrieves a variable value
