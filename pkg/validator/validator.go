package validator

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// DecodedString represents one known-good decoded string from Phase 5
type DecodedString struct {
	Index   int    `json:"index"`
	Content string `json:"content"`
	Pool    string `json:"pool_name,omitempty"`
}

// LoadPhase5Fixtures loads the 581 known-decoded strings from the ground truth file
func LoadPhase5Fixtures(filepath string) ([]DecodedString, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture file: %w", err)
	}

	var fixtures []DecodedString
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, fmt.Errorf("failed to parse fixtures: %w", err)
	}

	return fixtures, nil
}

// ValidateXORDecoding tests if a given XOR key produces the expected plaintext
func ValidateXORDecoding(ciphered []byte, key []byte, expected string) (bool, error) {
	if len(ciphered) != len(expected) {
		return false, fmt.Errorf("length mismatch: ciphered=%d, expected=%d", len(ciphered), len(expected))
	}

	// XOR decode
	decoded := make([]byte, len(ciphered))
	keyIdx := 0
	for i, b := range ciphered {
		if keyIdx >= len(key) {
			keyIdx = 0 // wrap around
		}
		decoded[i] = b ^ key[keyIdx]
		keyIdx++
	}

	// Compare with expected (as bytes)
	return string(decoded) == expected, nil
}

// BruteForceXORKey tests candidate keys against all fixtures
// Returns the key that decodes the most strings correctly
func BruteForceXORKey(candidateKeys [][]byte, fixtures []DecodedString, cipheredPool []byte) ([]byte, int, error) {
	bestKey := []byte{}
	bestScore := 0

	for _, key := range candidateKeys {
		score := 0
		for _, fixture := range fixtures {
			// Try decoding this fixture with this key
			valid, err := ValidateXORDecoding(cipheredPool, key, fixture.Content)
			if err == nil && valid {
				score++
			}
		}

		if score > bestScore {
			bestScore = score
			bestKey = key
		}

		if bestScore == len(fixtures) {
			break // Perfect match found
		}
	}

	return bestKey, bestScore, nil
}

// CribSearch tries to find the XOR key using known plaintext
// (crib-based attack, as Ocachs mentioned)
func CribSearch(cipheredPool []byte, knownPlaintext string) ([]byte, error) {
	if len(knownPlaintext) > len(cipheredPool) {
		return nil, fmt.Errorf("plaintext longer than ciphered pool")
	}

	// Try every possible position in the pool where the plaintext could be
	for offset := 0; offset <= len(cipheredPool)-len(knownPlaintext); offset++ {
		// Extract potential key from this position
		potentialKey := make([]byte, len(knownPlaintext))
		for i := 0; i < len(knownPlaintext); i++ {
			potentialKey[i] = cipheredPool[offset+i] ^ knownPlaintext[i]
		}

		// Validate this key against known fixtures
		// (This would need the actual ciphered pool and fixtures to test)
		return potentialKey, nil // Stub - return first candidate
	}

	return nil, fmt.Errorf("crib search failed")
}

// ReportValidationResults prints a summary of validation results
func ReportValidationResults(validCount int, totalCount int) {
	pct := float64(validCount) / float64(totalCount) * 100
	fmt.Printf("Validation: %d/%d strings decoded correctly (%.1f%%)\n", validCount, totalCount, pct)

	if validCount == totalCount {
		fmt.Println("✓ Perfect match! Decoder is fully validated.")
	} else if validCount >= totalCount*9/10 {
		fmt.Println("✓ Excellent. Minor decoding issues remain.")
	} else if validCount > 0 {
		fmt.Println("⚠ Partial match. Key or mechanism may be incorrect.")
	} else {
		fmt.Println("✗ No matches. Need to revise XOR mechanism or key calculation.")
	}
}
