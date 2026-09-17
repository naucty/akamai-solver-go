package main

import (
	"fmt"

	"github.com/naucty/akamai-solver-go/pkg/crypto"
)

func main() {
	fmt.Println("=== OC8 + RN8 Test ===\n")

	// Test OC8
	fmt.Println("[OC8] Testing LCG-based cipher")
	oc8 := crypto.NewOC8()

	plaintext := `{"ver":"1.0","fpt":123}`
	seed := uint32(0x12345678)

	encrypted := oc8.Encrypt(plaintext, seed)
	fmt.Printf("  Plain:     %s\n", plaintext)
	fmt.Printf("  Encrypted: %s\n", encrypted)

	oc8_2 := crypto.NewOC8()
	decrypted := oc8_2.Decrypt(encrypted, seed)
	fmt.Printf("  Decrypted: %s\n", decrypted)

	if decrypted == plaintext {
		fmt.Println("  ✓ Roundtrip OK\n")
	} else {
		fmt.Println("  ✗ Mismatch!\n")
	}

	// Test RN8
	fmt.Println("[RN8] Testing shuffle (seed=0x42)")
	text := "field_a:field_b:field_c:field_d"
	seed2 := uint32(0x42)

	shuffled := crypto.Brasser(text, seed2, ":")
	fmt.Printf("  Original:  %s\n", text)
	fmt.Printf("  Shuffled:  %s\n", shuffled)

	unshuffled := crypto.Debrasser(shuffled, seed2, ":")
	fmt.Printf("  Unshuffled:%s\n", unshuffled)

	if unshuffled == text {
		fmt.Println("  ✓ Roundtrip OK\n")
	} else {
		fmt.Println("  ✗ Mismatch!\n")
	}

	// Combined test
	fmt.Println("[Combined] OC8 + RN8 pipeline")
	original := "part1:part2:part3"
	rnSeed := uint32(0x11111111)
	ocSeed := uint32(0x22222222)

	// RN8 shuffle
	shuffled2 := crypto.Brasser(original, rnSeed, ":")
	fmt.Printf("  1. Original:  %s\n", original)
	fmt.Printf("  2. Shuffled:  %s\n", shuffled2)

	// OC8 encrypt
	oc8_3 := crypto.NewOC8()
	encrypted2 := oc8_3.Encrypt(shuffled2, ocSeed)
	fmt.Printf("  3. Encrypted: %s\n", encrypted2)

	// Reverse: OC8 decrypt
	oc8_4 := crypto.NewOC8()
	decrypted2 := oc8_4.Decrypt(encrypted2, ocSeed)
	fmt.Printf("  4. Decrypted: %s\n", decrypted2)

	// RN8 unshuffle
	unshuffled2 := crypto.Debrasser(decrypted2, rnSeed, ":")
	fmt.Printf("  5. Unshuffled:%s\n", unshuffled2)

	if unshuffled2 == original {
		fmt.Println("  ✓ Full pipeline OK\n")
	} else {
		fmt.Println("  ✗ Pipeline mismatch!\n")
	}
}
