package browser

import (
	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

// InitBrowserContext sets up global variables simulating a browser environment.
// Works with interp2 (AST-based interpreter).
func InitBrowserContext(scope *interp2.Scope) {
	// Global stacks and arrays for decoder execution
	scope.Declare("jI", []interp2.Value{})
	scope.Declare("wP", []interp2.Value{})

	// Counters and indices
	scope.Declare("NY", 0.0)
	scope.Declare("vc", 0.0)

	// Navigator mock
	navObj := &interp2.Object{Props: map[string]interp2.Value{
		"userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"language":  "en-US",
	}}
	scope.Declare("navigator", navObj)

	// Screen mock (pixel tracking)
	screenObj := &interp2.Object{Props: map[string]interp2.Value{
		"width":  1920.0,
		"height": 1080.0,
	}}
	scope.Declare("screen", screenObj)

	// Date.now() mock
	dateObj := &interp2.Object{Props: map[string]interp2.Value{
		"now": 1694961540000.0,
	}}
	scope.Declare("Date", dateObj)

	// window object (self-referential in browser; here just a reference to global scope)
	windowObj := &interp2.Object{Props: map[string]interp2.Value{
		"navigator": navObj,
		"screen":    screenObj,
	}}
	scope.Declare("window", windowObj)

	// Helper for tracking execution (sentry values)
	scope.Declare("kq", 0.0)
	scope.Declare("XI", -1.0)
	scope.Declare("kO", 0.0)
}

// SetPoolContext sets up pool-specific variables for a decoding run.
// This is called before executing the main tY function with a pool.
func SetPoolContext(scope *interp2.Scope, pool []byte, keystream []byte) {
	var poolVal interp2.Value = string(pool)
	var keystreamVal interp2.Value = string(keystream)

	scope.Declare("pV", poolVal)
	scope.Declare("sm", &interp2.Object{Props: map[string]interp2.Value{
		"Y8": keystreamVal,
	}})
}
