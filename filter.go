package main

import (
	"strings"
	"strconv"
	"fmt"
	"rand"
	"log"
	"os"
)

func PassesAllFilters(line JSONData, filters []Expression) (result bool, err os.Error) {
	for _, filter := range filters {
		passes, err := filter.Evaluate(line)
		if err != nil {
			return false, err
		}
		passes, ok := passes.(bool)
		if !ok {
			return false, fmt.Errorf("Expected a boolean for %T, got %T", filter, passes)
		}
		if !passes.(bool) {
			return false, nil
		}
	}
	return true, nil
}

/*
 * RandomSample(float64)
 *
 * Returns a boolean true with probability given by the first and only argument.
 */
type RandomSample struct {
	rate Expression
}

func (f *RandomSample) Setup(fname string, args []Expression) (err os.Error) {
	if len(args) != 1 {
		return fmt.Errorf("RandomSample takes a single argument, a float between 0 and 1")
	}
	f.rate = args[0]
	return
}

func (f *RandomSample) Evaluate(data JSONData) (result interface{}, err os.Error) {
	sampleRate, err := f.rate.Evaluate(data)
	if err != nil {
		return false, err
	}
	if sampleRate, ok := sampleRate.(float64); !ok {
		return false, fmt.Errorf("RandomSample takes a single argument, a float between 0 and 1. Got %v", sampleRate)
	}
	return rand.Float64() < sampleRate.(float64), nil
}

func (f *RandomSample) String() string {
	return fmt.Sprintf("RandomSample(%v)", f.rate)
}


/*
 * EveryNth(int)
 *
 * Returns a boolean true every nth time it's evaluated.
 */
type EveryNth struct {
	rate    Expression
	counter int
}

func (f *EveryNth) Setup(fname string, args []Expression) (err os.Error) {
	if len(args) != 1 {
		return fmt.Errorf("RandomSample takes a single argument, an int between 0 and 1")
	}
	f.rate = args[0]
	return
}

func (f *EveryNth) Evaluate(data JSONData) (result interface{}, err os.Error) {
	rate, err := f.rate.Evaluate(data)
	if err != nil {
		return false, err
	}
	if rate, ok := rate.(int); !ok {
		return false, fmt.Errorf("RandomSample takes a single argument, a positive integer. Got %v", rate)
	}
	f.counter++
	if f.counter >= rate.(int) {
		f.counter = 0
		return true, nil
	}
	return false, nil
}

func (f *EveryNth) String() string {
	return fmt.Sprintf("EveryNth(%v)", f.rate)
}

/*
 * Comparison Filter
 * 
 * XXX: Broken with the new parser. They need to be rewritten to the Expression interface.
 */
type ComparisonFilter struct {
	key      string
	operator func(a, b interface{}) bool
	rhs      interface{}
}

var comparisonOperators = map[string](func(a, b interface{}) bool){
	"==": func(a, b interface{}) bool { return a == b },
	"!=": func(a, b interface{}) bool { return a != b },
	">=": func(a, b interface{}) bool { return a.(float64) >= b.(float64) },
	"<=": func(a, b interface{}) bool { return a.(float64) <= b.(float64) },
	"<":  func(a, b interface{}) bool { return a.(float64) < b.(float64) },
	">":  func(a, b interface{}) bool { return a.(float64) > b.(float64) },
}

func NewComparisonFilter(query string) (f *ComparisonFilter, ok bool) {
	f = new(ComparisonFilter)
	ok, applicable, msg := f.Parse(query)
	if applicable {
		log.Print(msg)
	} else {
		ok = false
	}
	return
}

func (f *ComparisonFilter) Parse(query string) (ok bool, applicable bool, msg string) {
	fields := strings.Split(query, " ")
	applicable = true
	ok = true
	if len(fields) == 3 {
		key := fields[0]
		opStr := fields[1]
		rhs := fields[2]

		// Strip off quotes if there's a matching pair (e.g. allows checks for == "")
		if len(rhs) >= 2 && strings.HasPrefix(rhs, "\"") && strings.HasSuffix(rhs, "\"") {
			rhs = rhs[1 : len(rhs)-1]
		}

		operator, operatorPresent := comparisonOperators[opStr]
		if !operatorPresent {
			ok = false
			msg = "%v is not a valid operator."
			return
		} else {
			f.operator = operator
		}

		rhsFloat, err := strconv.Atof64(rhs)
		if err == nil {
			f.rhs = rhsFloat
		} else {
			f.rhs = rhs
		}

		f.key = key
	} else {
		applicable = false
		ok = false
	}
	return
}

func (f ComparisonFilter) Predicate(line JSONData) bool {
	lhs, ok := GetDeep(f.key, line)
	if ok {
		return f.operator(lhs, f.rhs)
	}
	return false
}
