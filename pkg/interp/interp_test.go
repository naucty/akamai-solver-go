package interp

import (
	"testing"
)

// TestCaseKX tests case KX (calculates Jt, MY from Bk context)
func TestCaseKX(t *testing.T) {
	interp := NewInterpreter()
	
	// First run case Dl to set base constants
	dlCode := `
var tP = + ! ![];
var nq = tP + tP;
var lV = tP + nq;
var ZU = lV + nq;
var hP = ZU * tP + nq;
var rm = lV + tP;
`
	if err := interp.ExecuteCase(dlCode); err != nil {
		t.Fatalf("Case Dl failed: %v", err)
	}
	
	// Now run case KX which depends on the above
	// From script: Jt=hP+ZU-rm*tP, MY=tP*ZU-nq+lV
	kxCode := `
var Jt = hP + ZU - rm * tP;
var MY = tP * ZU - nq + lV;
`
	if err := interp.ExecuteCase(kxCode); err != nil {
		t.Fatalf("Case KX failed: %v", err)
	}
	
	// Verify
	tests := []struct {
		name     string
		expected int64
	}{
		{"Jt", 8},  // 7+5-4*1 = 7+5-4 = 8
		{"MY", 6},  // 1*5-2+3 = 5-2+3 = 6
	}
	
	for _, tt := range tests {
		val, exists := interp.GetVar(tt.name)
		if !exists {
			t.Errorf("Variable %s not found", tt.name)
			continue
		}
		
		intVal, ok := val.(int64)
		if !ok {
			t.Errorf("Variable %s is not int64: %T", tt.name, val)
			continue
		}
		
		if intVal != tt.expected {
			t.Errorf("Variable %s: expected %d, got %d", tt.name, tt.expected, intVal)
		}
	}
	
	t.Logf("KX constants resolved: Jt=%v, MY=%v", 
		interp.Vars["Jt"], interp.Vars["MY"])
}

// TestCaseJW tests case jW (calculates OB and other final constants)
func TestCaseJW(t *testing.T) {
	interp := NewInterpreter()
	
	// Setup: Dl + KX + lj
	dlCode := `
var tP = + ! ![];
var nq = tP + tP;
var lV = tP + nq;
var ZU = lV + nq;
var hP = ZU * tP + nq;
var rm = lV + tP;
`
	if err := interp.ExecuteCase(dlCode); err != nil {
		t.Fatalf("Case Dl failed: %v", err)
	}
	
	kxCode := `
var Jt = hP + ZU - rm * tP;
var MY = tP * ZU - nq + lV;
`
	if err := interp.ExecuteCase(kxCode); err != nil {
		t.Fatalf("Case KX failed: %v", err)
	}
	
	// Also in KX: GL = hP*lV - MY*nq
	glCode := `
var GL = hP * lV - MY * nq;
`
	if err := interp.ExecuteCase(glCode); err != nil {
		t.Fatalf("Case KX (GL) failed: %v", err)
	}
	
	ljCode := `
var GV = hP + ZU * lV + nq + GL;
`
	if err := interp.ExecuteCase(ljCode); err != nil {
		t.Fatalf("Case lj failed: %v", err)
	}
	
	// Now case jW: OB=tP+Jt*rm+GV+MY
	jwCode := `
var OB = tP + Jt * rm + GV + MY;
`
	if err := interp.ExecuteCase(jwCode); err != nil {
		t.Fatalf("Case jW failed: %v", err)
	}
	
	val, exists := interp.GetVar("OB")
	if !exists {
		t.Fatalf("OB not found")
	}
	
	ob, ok := val.(int64)
	if !ok {
		t.Fatalf("OB is not int64: %T", val)
	}
	
	// OB = 1 + 8*4 + 33 + 6 = 1 + 32 + 33 + 6 = 72
	expected := int64(72)
	if ob != expected {
		t.Errorf("OB: expected %d, got %d", expected, ob)
	} else {
		t.Logf("OB calculated correctly: %d (wP[%d] is our XOR key stream)", ob, ob)
	}
}
