package main

import (
	"testing"
)

type samplingFilterTest struct {
	statement string
	rate      float64
	ok        bool
}

var samplingFilterTests = []samplingFilterTest{
	samplingFilterTest{"SAMPLE 0.5", 0.5, true},
	samplingFilterTest{"FOO 0.5", 0.0, false},
	samplingFilterTest{"SAMPLE 2", -1, false},
}

func TestParsingSamplingFilter(t *testing.T) {
	for _, test := range samplingFilterTests {
		f, ok := NewSamplingFilter(test.statement)
		if ok != test.ok {
			t.Errorf("For statement '%s', expected ok = %t, but was %t", test.statement, test.ok, ok)
		}
		if f.rate != test.rate {
			t.Errorf("For statement '%s', expected rate %d, but was %d", test.statement, test.rate, f.rate)
		}
	}
}
