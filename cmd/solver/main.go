package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/crypto"
	"github.com/naucty/akamai-solver-go/pkg/vm"
)

// SensorRequest contient les paramètres pour générer un sensor.
type SensorRequest struct {
	Bytecode string                 `json:"bytecode"` // hex-encoded bytecode
	Data     map[string]interface{} `json:"data"`
	OcSeed   uint32                 `json:"oc_seed"`
	RnSeed   uint32                 `json:"rn_seed"`
}

// SensorResponse retourne le sensor chiffré.
type SensorResponse struct {
	Sensor string `json:"sensor"`
	Fields int    `json:"fields_detected"`
	Error  string `json:"error,omitempty"`
}

// GenerateSensor: bytecode → ordre champs → JSON → RN8 → OC8 → sensor
func GenerateSensor(bytecode []byte, data map[string]interface{}, ocSeed, rnSeed uint32) (string, int, error) {
	// 1. Désassembler bytecode → ordre champs
	finder := &vm.StringFinder{Code: bytecode}
	ops := finder.FindStringOpcodes(8)
	opsMap := make(map[byte]bool)
	for _, o := range ops {
		opsMap[o] = true
	}
	instrs := finder.Disassemble(opsMap)
	fieldOrder := vm.ExtractStrings(instrs)
	normalized := vm.Normalize(fieldOrder)

	// 2. Construire JSON ordonné
	ordered := make(map[string]interface{})
	for _, field := range normalized {
		if val, ok := data[field]; ok {
			ordered[field] = val
		}
	}
	plainBytes, err := json.Marshal(ordered)
	if err != nil {
		return "", 0, fmt.Errorf("json marshal: %w", err)
	}
	plaintext := string(plainBytes)

	// 3. RN8 brassage
	shuffled := crypto.Brasser(plaintext, rnSeed, ":")

	// 4. OC8 chiffrement
	oc := crypto.NewOC8()
	sensor := oc.Encrypt(shuffled, ocSeed)

	return sensor, len(normalized), nil
}

// handleSensor HTTP handler
func handleSensor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req SensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(SensorResponse{Error: err.Error()})
		return
	}

	bytecode := []byte(req.Bytecode) // TODO: hex.DecodeString pour production

	sensor, fields, err := GenerateSensor(bytecode, req.Data, req.OcSeed, req.RnSeed)
	if err != nil {
		json.NewEncoder(w).Encode(SensorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(SensorResponse{
		Sensor: sensor,
		Fields: fields,
	})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/sensor", handleSensor)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	log.Printf("akamai-solver-go listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
