package crypto

// PRNG partagé par RN8 et OC8 (lu dans har_script.js).
const (
	mult uint32 = 65793
	add  uint32 = 4282663
	m32  uint32 = 0xFFFFFFFF
	m23  uint32 = 0x7FFFFF
)

func advance(k uint32) uint32 {
	return (((k * mult) & m32) + add) & m23
}

// alphabet: ASCII 32..126 sauf " (34), ' (39), \ (92) -> 92 caractères.
var alphabet = buildAlphabet()

func buildAlphabet() []byte {
	var a []byte
	for c := 32; c < 127; c++ {
		if c == 34 || c == 39 || c == 92 {
			continue
		}
		a = append(a, byte(c))
	}
	return a
}

// OC8 chiffre par décalage sur l'alphabet 92, flux dépendant d'une graine.
type OC8 struct {
	idx [256]int
}

func NewOC8() *OC8 {
	o := &OC8{}
	for i := range o.idx {
		o.idx[i] = -1
	}
	for i, c := range alphabet {
		o.idx[c] = i
	}
	return o
}

// Encrypt applique le chiffre. Le flux avance à CHAQUE position,
// même sur un caractère hors alphabet recopié tel quel.
func (o *OC8) Encrypt(text string, seed uint32) string {
	out := make([]byte, len(text))
	k := seed
	for i := 0; i < len(text); i++ {
		shift := int(((k >> 8) & 0xffff) % 92)
		k = advance(k)
		if j := o.idx[text[i]]; j >= 0 {
			out[i] = alphabet[(j+shift)%92]
		} else {
			out[i] = text[i]
		}
	}
	return string(out)
}

// Decrypt inverse Encrypt avec la même graine.
func (o *OC8) Decrypt(text string, seed uint32) string {
	out := make([]byte, len(text))
	k := seed
	for i := 0; i < len(text); i++ {
		shift := int(((k >> 8) & 0xffff) % 92)
		k = advance(k)
		if j := o.idx[text[i]]; j >= 0 {
			out[i] = alphabet[((j-shift)%92+92)%92]
		} else {
			out[i] = text[i]
		}
	}
	return string(out)
}
