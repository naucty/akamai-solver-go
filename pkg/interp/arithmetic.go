package interp

import (
	"fmt"
	"strconv"
	"strings"
)

// evaluateArithmeticExpr évalue une expression arithmétique avec respect de la préséance
// Gère: +, -, *, /, %, avec respect strict de l'ordre opérateur
func evaluateArithmeticExpr(expr string) Value {
	expr = strings.ReplaceAll(expr, " ", "")
	if expr == "" {
		return nil
	}
	
	result, err := parseAddExpr(expr)
	if err != nil {
		return nil
	}
	return result
}

// parseAddExpr handles + and - (lowest precedence)
func parseAddExpr(expr string) (int64, error) {
	// Find the last + or - that's not inside parentheses
	// Work right-to-left to handle left-associativity
	
	depth := 0
	var lastAddOp int = -1
	var lastAddChar rune
	
	for i := len(expr) - 1; i >= 0; i-- {
		ch := rune(expr[i])
		
		if ch == ')' {
			depth++
		} else if ch == '(' {
			depth--
		} else if depth == 0 && (ch == '+' || ch == '-') {
			// Make sure it's not a sign at the start or after another operator
			if i > 0 {
				prevChar := rune(expr[i-1])
				if prevChar != '+' && prevChar != '-' && prevChar != '*' && prevChar != '/' && prevChar != '(' {
					lastAddOp = i
					lastAddChar = ch
					break
				}
			}
		}
	}
	
	if lastAddOp >= 0 {
		left, err := parseAddExpr(expr[:lastAddOp])
		if err != nil {
			return 0, err
		}
		right, err := parseMulExpr(expr[lastAddOp+1:])
		if err != nil {
			return 0, err
		}
		
		if lastAddChar == '+' {
			return left + right, nil
		} else {
			return left - right, nil
		}
	}
	
	// No add/sub at this level, try multiplication
	return parseMulExpr(expr)
}

// parseMulExpr handles * and / (higher precedence than +/-)
func parseMulExpr(expr string) (int64, error) {
	depth := 0
	var lastMulOp int = -1
	var lastMulChar rune
	
	for i := len(expr) - 1; i >= 0; i-- {
		ch := rune(expr[i])
		
		if ch == ')' {
			depth++
		} else if ch == '(' {
			depth--
		} else if depth == 0 && (ch == '*' || ch == '/') {
			lastMulOp = i
			lastMulChar = ch
			break
		}
	}
	
	if lastMulOp >= 0 {
		left, err := parseMulExpr(expr[:lastMulOp])
		if err != nil {
			return 0, err
		}
		right, err := parseUnaryExpr(expr[lastMulOp+1:])
		if err != nil {
			return 0, err
		}
		
		if lastMulChar == '*' {
			return left * right, nil
		} else {
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		}
	}
	
	// No mul/div at this level, try unary/primary
	return parseUnaryExpr(expr)
}

// parseUnaryExpr handles unary operators and primary expressions
func parseUnaryExpr(expr string) (int64, error) {
	expr = strings.TrimSpace(expr)
	
	// Handle parentheses
	if strings.HasPrefix(expr, "(") && strings.HasSuffix(expr, ")") {
		return parseAddExpr(expr[1 : len(expr)-1])
	}
	
	// Handle negative numbers
	if strings.HasPrefix(expr, "-") {
		val, err := parseUnaryExpr(expr[1:])
		return -val, err
	}
	
	// Try to parse as number
	if num, err := strconv.ParseInt(expr, 10, 64); err == nil {
		return num, nil
	}
	
	return 0, fmt.Errorf("cannot parse: %s", expr)
}

// Also update evalExpr to use the new evaluator
func (i *Interpreter) evalArithmeticOld(exprStr string) (Value, error) {
	// The new version uses evaluateArithmeticExpr which handles precedence
	expr := exprStr
	
	// Replace known variables with their values
	for {
		changed := false
		for name, val := range i.Vars {
			if idx := strings.Index(expr, name); idx >= 0 {
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
	
	result := evaluateArithmeticExpr(expr)
	if result == nil {
		return nil, fmt.Errorf("unable to evaluate: %s", exprStr)
	}
	return result, nil
}
