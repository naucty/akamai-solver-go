package main

import (
	"fmt"

	"github.com/naucty/akamai-solver-go/pkg/interp2"
)

func main() {
	interp := interp2.NewInterpreter()

	code := `
	var fact = function(n) {
		if (n <= 1) { return 1; }
		return n * fact(n - 1);
	};
	var r = fact(6);

	var i = 0;
	var sum = 0;
	while (i < 5) {
		sum += i;
		i++;
	}

	var sw = 0;
	switch (2) {
		case 1:
			sw = 100;
			break;
		case 2:
			sw = 200;
			break;
		default:
			sw = 999;
	}

	function makeCounter() {
		var c = 0;
		return function() {
			c++;
			return c;
		};
	}
	var counter = makeCounter();
	var c1 = counter();
	var c2 = counter();
	var c3 = counter();
	`

	scope, err := interp.ParseAndRun(code)
	if err != nil {
		fmt.Println("ERROR:", err)
		return
	}

	for _, name := range []string{"r", "sum", "sw", "c1", "c2", "c3"} {
		v, _ := scope.Get(name)
		fmt.Printf("%s = %v\n", name, v)
	}
}
