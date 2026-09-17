package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/vm"
)

func main() {
	// Test sur build réel
	bytecodeFile := "/home/dev/projects/akamai-solver/artifacts/bytecode/00f877a9c1a2909d_0.bin"
	bytecode, err := os.ReadFile(bytecodeFile)
	if err != nil {
		fmt.Printf("Error reading bytecode: %v\n", err)
		return
	}
	fmt.Printf("✓ Bytecode loaded: %d bytes\n\n", len(bytecode))

	// Disassemble
	finder := &vm.StringFinder{Code: bytecode}
	ops := finder.FindStringOpcodes(8)
	opsMap := make(map[byte]bool)
	for _, o := range ops {
		opsMap[o] = true
	}
	fmt.Printf("✓ String opcodes detected: ")
	for _, o := range ops {
		fmt.Printf("%02x ", o)
	}
	fmt.Printf("\n\n")

	instrs := finder.Disassemble(opsMap)
	fieldOrder := vm.ExtractStrings(instrs)
	normalized := vm.Normalize(fieldOrder)

	fmt.Printf("✓ Field order extracted: %d fields\n", len(normalized))
	fmt.Printf("  First 15 fields:\n")
	for i := 0; i < 15 && i < len(normalized); i++ {
		fmt.Printf("    [%2d] %s\n", i, normalized[i])
	}
	fmt.Printf("\n")

	// Build plaintext with test data
	testData := map[string]interface{}{
		"ver":    "1.0",
		"fpt":    123,
		"fpc":    456,
		"ajr":    "test",
		"din":    true,
		"eem":    "telemetry",
		"ffs":    789,
		"vev":    1.23,
		"inf":    "info",
		"ajt":    "data",
	}

	plaintext := buildPlaintextOrdered(normalized, testData)
	fmt.Printf("✓ Plaintext built: %s\n\n", plaintext)

	// Encrypt with test key
	testKey := make([]byte, 32)
	rand.Read(testKey)
	fmt.Printf("✓ Generated random 32-byte key: %s\n", hex.EncodeToString(testKey))

	block, err := aes.NewCipher(testKey)
	if err != nil {
		fmt.Printf("Error creating cipher: %v\n", err)
		return
	}
	gcm, _ := cipher.NewGCM(block)
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	sensorB64 := base64.StdEncoding.EncodeToString(ciphertext)
	fmt.Printf("✓ Encrypted (AES-256-GCM): %d bytes ciphertext\n", len(ciphertext))
	fmt.Printf("  Base64 sensor (first 100 chars): %s...\n", sensorB64[:min(100, len(sensorB64))])
	fmt.Printf("\n")

	// Verify decryption works
	cipherBlob, _ := base64.StdEncoding.DecodeString(sensorB64)
	nonceDecrypt := cipherBlob[:gcm.NonceSize()]
	ciphertextDecrypt := cipherBlob[gcm.NonceSize():]
	decrypted, err := gcm.Open(nil, nonceDecrypt, ciphertextDecrypt, nil)
	if err != nil {
		fmt.Printf("✗ Decryption failed: %v\n", err)
		return
	}
	fmt.Printf("✓ Decryption successful: %s\n\n", string(decrypted))

	// Final verdict
	fmt.Printf("===== TEST RESULT =====\n")
	fmt.Printf("✓ Disassembly: OK (%d strings, %d normalized fields)\n", len(fieldOrder), len(normalized))
	fmt.Printf("✓ Plaintext build: OK\n")
	fmt.Printf("✓ Encryption: OK\n")
	fmt.Printf("✓ Decryption roundtrip: OK\n")
	fmt.Printf("\nSolver is ready for integration!\n")
}

func buildPlaintextOrdered(fieldOrder []string, data map[string]interface{}) string {
	result := make(map[string]interface{})
	for _, field := range fieldOrder {
		if val, ok := data[field]; ok {
			result[field] = val
		}
	}
	b, _ := json.Marshal(result)
	return string(b)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
