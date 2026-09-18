package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	scriptData, _ := os.ReadFile("/home/dev/projects/akamai-solver/artifacts/har_script.js")
	script := string(scriptData)

	interp := interp2.NewInterpreter()
	interp.InstallGlobals()

	// Capturer résultat sensor
	var sensorBytes []byte

	interp.Global.Set("__captureSensor__", interp2.NativeFunc(func(args []interp2.Value) (interp2.Value, error) {
		if len(args) > 0 {
			switch v := args[0].(type) {
			case string:
				sensorBytes = []byte(v)
				fmt.Printf("[SENSOR] Chaîne: %d octets\n", len(v))
			case float64:
				sensorBytes = []byte(fmt.Sprintf("%f", v))
			default:
				sensorBytes = []byte(fmt.Sprintf("%v", v))
			}
		}
		return nil, nil
	}))

	// Code wrapper: exécute le script, puis appelle la fonction sensor
	wrappedCode := script + `

// === WRAPPER CAPTEUR ===
// Appeler tY directement si accessible
if (typeof tY !== 'undefined') {
  var senseur = tY(typeof qG !== 'undefined' ? qG : 0, []);
  __captureSensor__(senseur);
} 
// Sinon chercher dans bmak
else if (typeof bmak !== 'undefined' && typeof bmak.get_sensor !== 'undefined') {
  var senseur = bmak.get_sensor();
  __captureSensor__(senseur);
} 
// Sinon chercher Eh
else if (typeof Eh !== 'undefined' && typeof HC !== 'undefined') {
  var senseur = Eh(HC, []);
  __captureSensor__(senseur);
} 
else {
  __captureSensor__("ERREUR: Aucune fonction trouvée");
}
// === FIN WRAPPER ===
`

	fmt.Println("Exécution script + wrapper...")
	_, err := interp.ParseAndRun(wrappedCode)
	if err != nil {
		fmt.Printf("Erreur: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n=== RÉSULTAT SENSOR ===")
	if len(sensorBytes) == 0 {
		fmt.Println("Aucun résultat capturé")
		os.Exit(1)
	}

	fmt.Printf("Octets bruts: %d\n", len(sensorBytes))
	fmt.Printf("Contenu: %s\n", string(sensorBytes)[:min(100, len(sensorBytes))])

	// Étape 2: Encoder en format sensor si nécessaire
	// Format attendu: "3;0;1;0;<seed>;<b64>;<metrics>;..."
	
	// Si le résultat est déjà en bytes, le mettre en base64
	b64Encoded := base64.StdEncoding.EncodeToString(sensorBytes)
	fmt.Printf("\nBase64: %s\n", b64Encoded)

	// Format complet (placeholder)
	sensorPayload := fmt.Sprintf("3;0;1;0;0;%s;0,0,0,0,0,0;", b64Encoded)
	fmt.Printf("\nPayload sensor: %s\n", sensorPayload)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
