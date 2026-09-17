package interp

import (
	"fmt"
	"strings"
)

// executeLoop handles while/do-while loops
func (i *Interpreter) handleLoop(stmt string) error {
	stmt = strings.TrimSpace(stmt)

	// do-while: do { ... } while(condition);
	if strings.HasPrefix(stmt, "do") {
		return i.executeDoWhile(stmt)
	}

	// while: while(condition) { ... }
	if strings.HasPrefix(stmt, "while") {
		return i.executeWhile(stmt)
	}

	// for: for(init; cond; inc) { ... }
	if strings.HasPrefix(stmt, "for") {
		return i.executeFor(stmt)
	}

	return fmt.Errorf("unknown loop type: %s", stmt[:20])
}

// executeWhile handles while(condition) { body }
func (i *Interpreter) executeWhile(stmt string) error {
	// Extract condition and body
	condStart := strings.Index(stmt, "(")
	if condStart == -1 {
		return fmt.Errorf("no opening paren in while")
	}

	condEnd := i.findMatchingParen(stmt, condStart)
	if condEnd == -1 {
		return fmt.Errorf("no closing paren for while condition")
	}

	condition := stmt[condStart+1 : condEnd]
	bodyStart := strings.Index(stmt[condEnd:], "{")
	if bodyStart == -1 {
		return fmt.Errorf("no body in while loop")
	}

	bodyStart += condEnd
	bodyEnd := i.findMatchingBrace(stmt, bodyStart)
	if bodyEnd == -1 {
		return fmt.Errorf("no closing brace for while body")
	}

	bodyCode := stmt[bodyStart+1 : bodyEnd]

	// Execute loop: evaluate condition, run body if true, repeat
	maxIterations := 10000 // prevent infinite loops
	iterations := 0

	for iterations < maxIterations {
		iterations++

		// Evaluate condition
		condVal, err := i.evalExpr(condition)
		if err != nil {
			return fmt.Errorf("while condition eval failed: %w", err)
		}

		// Check if condition is truthy
		if !i.isTruthy(condVal) {
			break
		}

		// Execute body
		savedCode := i.Code
		i.Code = bodyCode
		if err := i.execute(); err != nil && !strings.Contains(err.Error(), "break") {
			i.Code = savedCode
			return err
		}
		i.Code = savedCode
	}

	if iterations >= maxIterations {
		return fmt.Errorf("while loop exceeded max iterations")
	}

	return nil
}

// executeDoWhile handles do { body } while(condition);
func (i *Interpreter) executeDoWhile(stmt string) error {
	// Extract body and condition
	bodyStart := strings.Index(stmt, "{")
	if bodyStart == -1 {
		return fmt.Errorf("no opening brace in do-while")
	}

	bodyEnd := i.findMatchingBrace(stmt, bodyStart)
	if bodyEnd == -1 {
		return fmt.Errorf("no closing brace for do-while body")
	}

	bodyCode := stmt[bodyStart+1 : bodyEnd]

	// Find while(condition)
	whileStart := strings.Index(stmt[bodyEnd:], "while")
	if whileStart == -1 {
		return fmt.Errorf("no while in do-while")
	}
	whileStart += bodyEnd

	condStart := strings.Index(stmt[whileStart:], "(")
	if condStart == -1 {
		return fmt.Errorf("no paren in do-while condition")
	}
	condStart += whileStart

	condEnd := i.findMatchingParen(stmt, condStart)
	if condEnd == -1 {
		return fmt.Errorf("no closing paren for do-while condition")
	}

	condition := stmt[condStart+1 : condEnd]

	// Execute: run body once, then loop while condition is true
	maxIterations := 10000
	iterations := 0

	for iterations < maxIterations {
		iterations++

		// Execute body
		savedCode := i.Code
		i.Code = bodyCode
		if err := i.execute(); err != nil && !strings.Contains(err.Error(), "break") {
			i.Code = savedCode
			return err
		}
		i.Code = savedCode

		// Evaluate condition
		condVal, err := i.evalExpr(condition)
		if err != nil {
			return fmt.Errorf("do-while condition eval failed: %w", err)
		}

		// Check if condition is truthy
		if !i.isTruthy(condVal) {
			break
		}
	}

	if iterations >= maxIterations {
		return fmt.Errorf("do-while loop exceeded max iterations")
	}

	return nil
}

// executeFor handles for(init; cond; inc) { body }
func (i *Interpreter) executeFor(stmt string) error {
	// Extract the for(...) part
	parStart := strings.Index(stmt, "(")
	if parStart == -1 {
		return fmt.Errorf("no opening paren in for")
	}

	parEnd := i.findMatchingParen(stmt, parStart)
	if parEnd == -1 {
		return fmt.Errorf("no closing paren in for")
	}

	forHeader := stmt[parStart+1 : parEnd]

	// Split init, cond, inc by semicolon (careful with nested expressions)
	parts := i.splitForHeader(forHeader)
	if len(parts) < 3 {
		return fmt.Errorf("invalid for header: expected 3 parts, got %d", len(parts))
	}

	init := strings.TrimSpace(parts[0])
	cond := strings.TrimSpace(parts[1])
	inc := strings.TrimSpace(parts[2])

	// Extract body
	bodyStart := strings.Index(stmt[parEnd:], "{")
	if bodyStart == -1 {
		return fmt.Errorf("no body in for loop")
	}
	bodyStart += parEnd

	bodyEnd := i.findMatchingBrace(stmt, bodyStart)
	if bodyEnd == -1 {
		return fmt.Errorf("no closing brace for for body")
	}

	bodyCode := stmt[bodyStart+1 : bodyEnd]

	// Execute: init, then loop
	if init != "" {
		if err := i.executeStatement(init); err != nil {
			return err
		}
	}

	maxIterations := 10000
	iterations := 0

	for iterations < maxIterations {
		iterations++

		// Check condition
		if cond != "" {
			condVal, err := i.evalExpr(cond)
			if err != nil {
				return err
			}
			if !i.isTruthy(condVal) {
				break
			}
		}

		// Execute body
		savedCode := i.Code
		i.Code = bodyCode
		if err := i.execute(); err != nil && !strings.Contains(err.Error(), "break") {
			i.Code = savedCode
			return err
		}
		i.Code = savedCode

		// Execute increment
		if inc != "" {
			if err := i.executeStatement(inc); err != nil {
				return err
			}
		}
	}

	if iterations >= maxIterations {
		return fmt.Errorf("for loop exceeded max iterations")
	}

	return nil
}

// isTruthy evaluates JS truthiness
func (i *Interpreter) isTruthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case float64:
		return v != 0
	case string:
		return v != ""
	case nil:
		return false
	default:
		return true
	}
}

// findMatchingParen finds the closing paren for an opening paren
func (i *Interpreter) findMatchingParen(s string, openPos int) int {
	if openPos < 0 || openPos >= len(s) || s[openPos] != '(' {
		return -1
	}

	depth := 1
	for j := openPos + 1; j < len(s); j++ {
		if s[j] == '(' {
			depth++
		} else if s[j] == ')' {
			depth--
			if depth == 0 {
				return j
			}
		}
	}

	return -1
}

// findMatchingBrace finds the closing brace for an opening brace
func (i *Interpreter) findMatchingBrace(s string, openPos int) int {
	if openPos < 0 || openPos >= len(s) || s[openPos] != '{' {
		return -1
	}

	depth := 1
	for j := openPos + 1; j < len(s); j++ {
		if s[j] == '{' {
			depth++
		} else if s[j] == '}' {
			depth--
			if depth == 0 {
				return j
			}
		}
	}

	return -1
}

// splitForHeader splits "init; cond; inc" by semicolon
func (i *Interpreter) splitForHeader(header string) []string {
	var parts []string
	var current strings.Builder
	depth := 0

	for _, ch := range header {
		if ch == '(' || ch == '[' || ch == '{' {
			depth++
		} else if ch == ')' || ch == ']' || ch == '}' {
			depth--
		} else if ch == ';' && depth == 0 {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		current.WriteRune(ch)
	}

	parts = append(parts, current.String())
	return parts
}
