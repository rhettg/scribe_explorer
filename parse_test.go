package main

import (
	"testing"
	"fmt"
	"os"
)

type parseStringTest struct {
	statement string
	fname     string
	args      []string
	ok        bool
}

var parseStringTests = []parseStringTest{
	parseStringTest{"Foo(a", "", []string{}, false},
	parseStringTest{"Foo(1,2)", "Foo", []string{"1", "2"}, true},
	parseStringTest{"Subtract(a)", "Subtract", []string{"a"}, true},
	parseStringTest{"Subtract(a,b)", "Subtract", []string{"a", "b"}, true},
	parseStringTest{"Subtract(a,b,c)", "Subtract", []string{"a", "b", "c"}, true},
	parseStringTest{"Foo(Bar(abc),bc)", "Foo", []string{"Bar(abc)", "bc"}, true},
	parseStringTest{"Foo(Bar(a,b),c,de)", "Foo", []string{"Bar(a,b)", "c", "de"}, true},
	parseStringTest{"Foo(a,Bar(b,c)", "", []string{}, false}, // Unbalanced parens
	parseStringTest{"foo", "", []string{}, false},
}

func TestParseFunction(t *testing.T) {
	for _, test := range parseStringTests {
		fname, args, err := ParseString(test.statement)
		if test.ok && err != nil {
			t.Errorf("For statement '%s', expected nil err, but was %v", test.statement, err)
		}

		if !test.ok && err == nil {
			t.Errorf("For statement '%s', expected err, but was nil", test.statement)
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
