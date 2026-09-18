package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

type SensorRequest struct {
	Script   string                 `json:"script"`
	Navigator map[string]interface{} `json:"navigator"`
	Window   map[string]interface{} `json:"window"`
	BmSz     string                 `json:"bm_sz"`
	UserAgent string                `json:"user_agent"`
	URL      string                 `json:"url"`
}

type SensorResponse struct {
	Sensor string `json:"sensor"`
	Error  string `json:"error,omitempty"`
}

func handleSensor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, _ := io.ReadAll(r.Body)
	var req SensorRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SensorResponse{Error: err.Error()})
		return
	}

	// Execute script + generate sensor
	sensor, err := generateSensor(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SensorResponse{Error: err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SensorResponse{Sensor: sensor})
}

func generateSensor(req SensorRequest) (string, error) {
	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	// 1. Inject globals (mock browser environment)
	// bm_sz is what browser sends
	interp.Global.Set("bm_sz", req.BmSz)
	
	// 2. Set navigator object
	navObj := &interp2.Object{Props: make(map[string]interp2.Value)}
	navObj.Props["userAgent"] = req.UserAgent
	interp.Global.Set("navigator", navObj)

	// 3. Add export wrapper to capture hs() calls
	// Append export code to script
	exportCode := `
var _SENSOR_EXPORT = null;
var _HS_ORIGINAL = typeof hs !== 'undefined' ? hs : null;
if (_HS_ORIGINAL) {
  hs = function() {
    var result = _HS_ORIGINAL.apply(this, arguments);
    _SENSOR_EXPORT = result;
    return result;
  };
}
`
	fullScript := req.Script + "\n" + exportCode

	// 4. Execute
	_, err := interp.ParseAndRun(fullScript)
	if err != nil {
		return "", fmt.Errorf("script exec: %w", err)
	}

	// 5. Try to call hs() with params
	hsVal, ok := interp.Global.Get("hs")
	if !ok || hsVal == nil {
		// Try to find it in bmak
		bmak, ok := interp.Global.Get("bmak")
		if !ok || bmak == nil {
			return "", fmt.Errorf("hs not found in global scope")
		}
		// bmak.get_sensor or similar
		return "", fmt.Errorf("hs wrapped in bmak, not yet implemented")
	}

	// 6. Get the export value that was captured
	exportVal, ok := interp.Global.Get("_SENSOR_EXPORT")
	if !ok || exportVal == nil {
		return "", fmt.Errorf("_SENSOR_EXPORT not captured")
	}

	// 7. Encode to sensor format
	// TODO: Convert exportVal to "3;0;1;0;seed;b64;metrics;..."
	
	return fmt.Sprintf("3;0;1;0;0;%v;0,0,0,0,0,0;", exportVal), nil
}


func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/api/sensor", handleSensor)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("Sensor API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
