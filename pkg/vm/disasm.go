package vm

import (
	"fmt"
	"regexp"
	"sort"
)

// StringFinder detects string-carrying opcodes by shape analysis.
type StringFinder struct {
	Code []byte
}

// FindStringOpcodes returns bytes that behave like "OP u16_len <printable_payload>".
// Criteria: by form, no hardcoding. Returns up to 2 candidates (format has exactly 2).
func (sf *StringFinder) FindStringOpcodes(minSuccesses int) []byte {
	success := make(map[byte]int)

	for i := 0; i < len(sf.Code)-3; i++ {
		ln := int(sf.Code[i+1])<<8 | int(sf.Code[i+2])
		if ln < 1 || ln > 64 {
			continue
		}
		end := i + 3 + ln
		if end > len(sf.Code) {
			continue
		}
		payload := sf.Code[i+3 : end]
		allPrintable := true
		for _, c := range payload {
			if c < 32 || c >= 127 {
				allPrintable = false
				break
			}
		}
		if allPrintable {
			success[sf.Code[i]]++
		}
	}

	type kv struct {
		op byte
		n  int
	}
	var sorted []kv
	for op, n := range success {
		if n >= minSuccesses {
			sorted = append(sorted, kv{op, n})
		}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].n > sorted[j].n })

	result := make([]byte, 0, 2)
	for i := 0; i < 2 && i < len(sorted); i++ {
		result = append(result, sorted[i].op)
	}
	return result
}

// Instruction is a decoded bytecode instruction.
type Instruction struct {
	Offset  int
	Op      byte
	Payload string // non-empty if string-carrying
}

// Disassemble walks bytecode linearly. No hardcoded opcodes.
func (sf *StringFinder) Disassemble(stringOps map[byte]bool) []Instruction {
	var instrs []Instruction
	i := 0
	for i < len(sf.Code) {
		op := sf.Code[i]
		offset := i
		if stringOps[op] && i+3 <= len(sf.Code) {
			ln := int(sf.Code[i+1])<<8 | int(sf.Code[i+2])
			end := i + 3 + ln
			if end <= len(sf.Code) && ln > 0 {
				payload := sf.Code[i+3 : end]
				allPrintable := true
				for _, c := range payload {
					if c < 32 || c >= 127 {
						allPrintable = false
						break
					}
				}
				if allPrintable {
					instrs = append(instrs, Instruction{Offset: offset, Op: op, Payload: string(payload)})
					i = end
					continue
				}
			}
		}
		instrs = append(instrs, Instruction{Offset: offset, Op: op})
		i++
	}
	return instrs
}

// Normalize renumbers internal identifiers by first-use order.
// Keeps JS API names verbatim ("window", "navigator"...).
func Normalize(strs []string) []string {
	jsAPI := map[string]bool{
		"window": true, "navigator": true, "userAgent": true, "split": true,
		"join": true, "toString": true, "length": true, "charCodeAt": true,
		"String": true, "slice": true, "bmak": true, "startTs": true,
		"push": true, "Array": true, "isArray": true, "indexOf": true,
		"document": true, "createElement": true, "Math": true, "random": true,
		"call": true, "apply": true, "concat": true, "substring": true,
		"replace": true, "parseInt": true, "JSON": true, "stringify": true,
		"charAt": true, "fromCharCode": true, "Date": true, "getTime": true,
		"now": true, "floor": true,
	}
	re := regexp.MustCompile(`^[A-Za-z0-9]{1,2}$`)
	corr := make(map[string]string)
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		if !jsAPI[s] && re.MatchString(s) {
			if _, ok := corr[s]; !ok {
				corr[s] = fmt.Sprintf("v%d", len(corr))
			}
			result = append(result, corr[s])
		} else {
			result = append(result, s)
		}
	}
	return result
}

// ExtractStrings returns all string payloads in program order.
func ExtractStrings(instrs []Instruction) []string {
	result := make([]string, 0)
	for _, in := range instrs {
		if in.Payload != "" {
			result = append(result, in.Payload)
		}
	}
	return result
}
