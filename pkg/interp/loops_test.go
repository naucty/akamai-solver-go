package interp

import (
	"testing"
)

// TestWhileLoop tests basic while loop execution
func TestWhileLoop(t *testing.T) {
	interp := NewInterpreter()

	// Simple test: while(i < 5) { i++ }
	code := `
var i = 0;
var sum = 0;
while(i < 5) {
  sum = sum + i;
  i = i + 1;
}
`

	if err := interp.ExecuteCase(code); err != nil {
		t.Fatalf("Loop execution failed: %v", err)
	}

	// Check final values
	i, ok := interp.GetVar("i")
	if !ok || i != int64(5) {
		t.Errorf("Expected i=5, got %v", i)
	}

	sum, ok := interp.GetVar("sum")
	if !ok || sum != int64(10) {
		t.Errorf("Expected sum=10 (0+1+2+3+4), got %v", sum)
	}
}

// TestDoWhileLoop tests do-while loop
func TestDoWhileLoop(t *testing.T) {
	interp := NewInterpreter()

	code := `
var x = 10;
var count = 0;
do {
  count = count + 1;
  x = x - 1;
} while(x > 5);
`

	if err := interp.ExecuteCase(code); err != nil {
		t.Fatalf("Do-while failed: %v", err)
	}

	x, ok := interp.GetVar("x")
	if !ok || x != int64(5) {
		t.Errorf("Expected x=5, got %v", x)
	}

	count, ok := interp.GetVar("count")
	if !ok || count != int64(5) {
		t.Errorf("Expected count=5, got %v", count)
	}
}

// TestForLoop tests for loop
func TestForLoop(t *testing.T) {
	interp := NewInterpreter()

	code := `
var total = 0;
for(var j = 0; j < 6; j = j + 1) {
  total = total + j;
}
`

	if err := interp.ExecuteCase(code); err != nil {
		t.Fatalf("For loop failed: %v", err)
	}

	total, ok := interp.GetVar("total")
	if !ok || total != int64(15) {
		t.Errorf("Expected total=15 (0+1+2+3+4+5), got %v", total)
	}
}

// TestLoopFromActualCase tests pattern from real case cp
// case cp: while(QY(vc, pV.length)){...}
// Here QY = (a, b) => a < b, so while(vc < pV.length)
func TestLoopFromActualCase(t *testing.T) {
	interp := NewInterpreter()

	// Pre-set variables as they would be in real execution
	interp.SetVar("pV", []interface{}{'a', 'b', 'c', 'd', 'e'}) // 5-char array
	interp.SetVar("vc", int64(0))                                 // start
	interp.SetVar("NY", int64(0))                                 // keystream index
	interp.SetVar("Pd", "")                                       // accumulated output

	// Stub QY function: (a < b)
	// For now, test just the loop counter behavior
	code := `
while(vc < 5) {
  vc = vc + 1;
}
`

	if err := interp.ExecuteCase(code); err != nil {
		t.Fatalf("Case cp-like loop failed: %v", err)
	}

	vc, ok := interp.GetVar("vc")
	if !ok || vc != int64(5) {
		t.Errorf("Expected vc=5, got %v", vc)
	}
}
