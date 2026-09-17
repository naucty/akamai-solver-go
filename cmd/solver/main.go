package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/vm"
)

// SensorRequest contains parameters for sensor generation
type SensorRequest struct {
	Script string            `json:"script"`
	Data   map[string]interface{} `json:"data"`
}

// SensorResponse returns generated sensor
type SensorResponse struct {
	Sensor string `json:"sensor"`
	Error  string `json:"error,omitempty"`
}

// Solver holds the main logic
type Solver struct {
	encryptionKey []byte
}

// NewSolver initializes solver with encryption key
func NewSolver(keyBase64 string) (*Solver, error) {
	key, err := base64.StdEncoding.DecodeString(keyBase64)
	if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes")
	}
	return &Solver{encryptionKey: key}, nil
}

// GenerateSensor takes script and data, returns encrypted sensor
func (s *Solver) GenerateSensor(scriptCode string, data map[string]interface{}) (string, error) {
	// 1. Extract bytecode from script
	bytecode, err := extractBytecodeFromScript(scriptCode)
	if err != nil {
		return "", fmt.Errorf("extract bytecode: %w", err)
	}

	// 2. Disassemble to get field order
	finder := &vm.StringFinder{Code: bytecode}
	ops := finder.FindStringOpcodes(8)
	opsMap := make(map[byte]bool)
	for _, o := range ops {
		opsMap[o] = true
	}
	instrs := finder.Disassemble(opsMap)
	fieldOrder := vm.ExtractStrings(instrs)
	normalized := vm.Normalize(fieldOrder)

	// 3. Build plaintext JSON in order
	plaintext := buildPlaintext(normalized, data)

	// 4. Encrypt
	sensor, err := s.encryptSensor(plaintext)
	if err != nil {
		return "", fmt.Errorf("encrypt: %w", err)
	}

	return sensor, nil
}

// extractBytecodeFromScript extracts bytecode blob from script
// (stub: in real impl, parse and extract base64/hex blob)
func extractBytecodeFromScript(script string) ([]byte, error) {
	// TODO: implement proper extraction
	// For now, assume input is raw bytecode (for testing)
	return []byte(script), nil
}

// buildPlaintext builds JSON with fields in specified order
func buildPlaintext(fieldOrder []string, data map[string]interface{}) string {
	result := make(map[string]interface{})
	for _, field := range fieldOrder {
		if val, ok := data[field]; ok {
			result[field] = val
		}
	}
	b, _ := json.Marshal(result)
	return string(b)
}

// encryptSensor encrypts plaintext with AES-256-GCM
func (s *Solver) encryptSensor(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// handleSensor is HTTP handler
func (s *Solver) handleSensor(w http.ResponseWriter, r *http.Request) {
	var req SensorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(SensorResponse{Error: err.Error()})
		return
	}

	sensor, err := s.GenerateSensor(req.Script, req.Data)
	if err != nil {
		json.NewEncoder(w).Encode(SensorResponse{Error: err.Error()})
		return
	}

	json.NewEncoder(w).Encode(SensorResponse{Sensor: sensor})
}

func main() {
	// Initialize solver
	// TODO: load key from env or config
	keyBase64 := os.Getenv("SENSOR_KEY")
	if keyBase64 == "" {
		keyBase64 = "dGVzdGtleWZvcmVuY3J5cHRpb24xMjM0NTY3ODk=" // base64 dummy key
	}

	solver, err := NewSolver(keyBase64)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/sensor", solver.handleSensor)
	log.Printf("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
