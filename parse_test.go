package main

import (
	"testing"
	"fmt"
	"os"
)

type parseFunctionTest struct {
	statement string
	fname string
	args []string
	err os.Error 
}

var parseFunctionTests = []parseFunctionTest {
	parseFunctionTest{"Subtract(a)", "Subtract", []string{"a",}, nil},
	parseFunctionTest{"Subtract(a,b)", "Subtract", []string{"a", "b"}, nil},
}

func TestParseFunction(t *testing.T) {
	for _, test := range parseFunctionTests {
		fname, args, err := ParseFunction(test.statement)
		if err != test.err {
			t.Errorf("For statement '%s', expected err = %v, but was %v", test.statement, test.err, err)
		}

		if fname != test.fname {
			t.Errorf("For statement '%s', expected fname = %v, but was %v", test.statement, test.fname, aggregator)
		}
		
		if ok, err := sliceEquals(args, test.args); !ok {
			t.Errorf("For statement '%s', expected args = %v, but was %v: %s", test.statement, test.args, args, err)
		}
	}
}

func sliceEquals(x, y []string) (ok bool, err os.Error) {
	if len(x) != len(y) {
		return false, fmt.Errorf("Error: len(x) %d != len(y) %d", len(x), len(y)) 
	}
	for i, _ := range x {
		if x[i] != y[i] {
			return false, fmt.Errorf("Error: x[%d], %v != y[%d], %v", i, x[i], i, y[i])
		}
	}
	return true, nil
}
