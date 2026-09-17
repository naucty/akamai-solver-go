package crypto

import "strings"

// RN8 brasse les parts d'un texte (séparateur ':') avec un PRNG LCG.
// Algorithme: lu dans har_script.js @506058 (case Tg).

func rn8Pairs(n int, seed uint32) [][2]int {
	k := seed
	pairs := make([][2]int, n)
	for i := 0; i < n; i++ {
		iA := int(((k >> 8) & 0xffff)) % n
		k = advance(k)
		iB := int(((k >> 8) & 0xffff)) % n
		k = advance(k)
		pairs[i] = [2]int{iA, iB}
	}
	return pairs
}

// Brasser applique le brassage RN8 (sens aller = comme le script Akamai).
func Brasser(text string, seed uint32, sep string) string {
	if sep == "" {
		sep = ":"
	}
	parts := strings.Split(text, sep)
	n := len(parts)
	if n < 2 {
		return text
	}
	for _, p := range rn8Pairs(n, seed) {
		parts[p[0]], parts[p[1]] = parts[p[1]], parts[p[0]]
	}
	return strings.Join(parts, sep)
}

// Debrasser inverse le brassage RN8 (rejoue les échanges à l'envers).
func Debrasser(text string, seed uint32, sep string) string {
	if sep == "" {
		sep = ":"
	}
	parts := strings.Split(text, sep)
	n := len(parts)
	if n < 2 {
		return text
	}
	pairs := rn8Pairs(n, seed)
	for i := len(pairs) - 1; i >= 0; i-- {
		p := pairs[i]
		parts[p[0]], parts[p[1]] = parts[p[1]], parts[p[0]]
	}
	return strings.Join(parts, sep)
}
