package main

import (
	"testing"
)

type parseParameterTest struct {
	statement string
	parameters []string
	aggregator string
	ok bool
}

var parseParametersTests = []parseParameterTest {
	parseParameterTest{"Timer a b", []string{"a", "b"}, "Timer", true},
	parseParameterTest{"Average a 10", []string{"a", "10"}, "Average", true},
}

func TestParseParameters(t *testing.T) {
	for _, test := range parseParametersTests {
		aggregator, parameters, ok := ParseParameters(test.statement)
		if ok != test.ok {
			t.Errorf("For statement '%s', expected ok = %t, but was %t", test.statement, test.ok, ok)
		}

		if aggregator != test.aggregator {
			t.Errorf("For statement'%s', expected aggregator = %v, but was %v", test.statement, test.aggregator, aggregator)
		}
		
		t.Logf("Parameters %v =?= %v", test.parameters, parameters)
		/*
		for i, _ := range test.parameters {
			if parameters[i] != test.parameters[i] {
				t.Errorf("For statement'%s', expected params = %v, but was %v", test.statement, test.parameters, parameters)
			}
		}
		*/
	}
}
