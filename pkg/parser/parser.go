package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// Script represents Akamai script metadata
type Script struct {
	Hash    string
	Content string
	Raw     []byte
}

// BytecodeInfo holds decoded bytecode info
type BytecodeInfo struct {
	Opcodes []byte
	Strings []string
	Pools   map[string]string
}

// ExtractScriptFromHTML extracts Akamai script from HTML
func ExtractScriptFromHTML(html string) (string, error) {
	// Pattern: <script ... src=".../_bm...js" or inline script
	patterns := []string{
		`<script[^>]*src="([^"]*/_bm[^"]*)"`,
		`<script[^>]*>(.*?bmak.*?)</script>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(html)
		if len(matches) > 1 {
			return matches[1], nil
		}
	}
	return "", fmt.Errorf("akamai script not found")
}

// ExtractBytecodeFromScript extracts bytecode pool from script
func ExtractBytecodeFromScript(script string) ([]byte, error) {
	// Look for bytecode pattern: encoded hex or base64 blob
	// Common pattern: bmak blob with push() operations
	
	// Extract opcodes region - typically after initial declarations
	re := regexp.MustCompile(`\[\s*(\d+(?:,\s*\d+)*)\s*\]`)
	matches := re.FindStringSubmatch(script)
	if len(matches) > 0 {
		// Parse opcode array
		opsStr := strings.ReplaceAll(matches[1], " ", "")
		opStrs := strings.Split(opsStr, ",")
		opcodes := make([]byte, len(opStrs))
		for i, s := range opStrs {
			var op byte
			fmt.Sscanf(s, "%d", &op)
			opcodes[i] = op
		}
		return opcodes, nil
	}
	return nil, fmt.Errorf("bytecode not found")
}

// ExtractXORPools extracts XOR-encoded pools from script
func ExtractXORPools(script string) (map[string]string, error) {
	pools := make(map[string]string)
	
	// Pattern: decode pools with known XOR keys
	// Phase 5 identified key: look for patterns like bmak.pool = {key: xorvalue}
	
	re := regexp.MustCompile(`(\w+):\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(script, -1)
	
	for _, match := range matches {
		pools[match[1]] = match[2]
	}
	
	return pools, nil
}

// DecodeXORPool decodes a XOR-encoded pool with given key
func DecodeXORPool(encoded string, key byte) string {
	result := make([]byte, len(encoded))
	for i := 0; i < len(encoded); i++ {
		result[i] = encoded[i] ^ key
	}
	return string(result)
}

// ExtractFieldNames extracts sensor field names from script
func ExtractFieldNames(script string) ([]string, error) {
	// Pattern: look for field name declarations (pnte, qs, st, etc)
	// These appear as function names or object keys
	
	re := regexp.MustCompile(`\b([a-z]{2,4})\s*[:=]\s*function|\b([a-z]{2,4})\s*{`)
	matches := re.FindAllStringSubmatch(script, -1)
	
	fieldSet := make(map[string]bool)
	for _, match := range matches {
		if match[1] != "" {
			fieldSet[match[1]] = true
		} else if match[2] != "" {
			fieldSet[match[2]] = true
		}
	}
	
	fields := make([]string, 0, len(fieldSet))
	for f := range fieldSet {
		if len(f) >= 2 && len(f) <= 4 {
			fields = append(fields, f)
		}
	}
	
	return fields, nil
}
