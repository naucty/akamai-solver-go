package browser

import (
	"github.com/naucty/akamai-solver-go/pkg/interp"
)

// InitBrowserContext sets up global variables simulating a browser environment
// This is used to mock window, navigator, etc. before executing decoder cases
func InitBrowserContext(interp *interp.Interpreter) {
	// Global stacks and arrays used by the VM
	interp.SetVar("jI", []interface{}{})         // Stack for decoder state
	interp.SetVar("kq", int64(0))                // constant zero (used in many places)
	interp.SetVar("QY", int64(0))                // Comparison function: QY(a, b) => a < b
	
	// XOR keystream (will be set dynamically per pool)
	interp.SetVar("wP", []interface{}{})         // XOR keystream array
	
	// Global decoder state
	interp.SetVar("NY", int64(0))                // keystream position
	interp.SetVar("vc", int64(0))                // current pool index
	
	// Navigator-like properties
	navObj := map[string]interface{}{
		"userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"language":  "en-US",
	}
	interp.SetVar("navigator", navObj)
	
	// Window.screen properties (pixel tracking)
	screenObj := map[string]interface{}{
		"width":  int64(1920),
		"height": int64(1080),
	}
	interp.SetVar("screen", screenObj)
	
	// Fake timestamp
	interp.SetVar("Date", map[string]interface{}{
		"now": int64(1694961540000), // fixed timestamp for reproducibility
	})
	
	// Common helper constants
	interp.SetVar("kO", int64(0))   // goto label (will be set per dispatcher)
	interp.SetVar("XI", int64(-1))  // error sentinel
}

// SetPoolContext sets up the pool variables for a specific decoding run
// This is called before executing the main decoding loop
func SetPoolContext(interp *interp.Interpreter, pool []byte, keystream []byte) {
	// Convert bytes to character arrays (as JS would see them)
	poolChars := make([]interface{}, len(pool))
	for i, b := range pool {
		poolChars[i] = int64(b)
	}
	interp.SetVar("pV", poolChars) // The encrypted pool
	
	keystreamChars := make([]interface{}, len(keystream))
	for i, b := range keystream {
		keystreamChars[i] = int64(b)
	}
	interp.SetVar("wP", keystreamChars) // The keystream
	
	// Reset loop counters
	interp.SetVar("vc", int64(0))
	interp.SetVar("NY", int64(0))
}
