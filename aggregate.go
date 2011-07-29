package main

import (
	"strings"
	"strconv"
	"log"
	"fmt"
	"container/list"
)

type Aggregator interface {
	Parse(params []string, query string) (msg string, ok bool)
	Push(element JSONData) interface{}
	String() string
}

func ParseAggregatorStatement(query string) (a Aggregator) {
	aggType, params, ok := ParseParameters(query)
	if !ok {
		return nil
	}
	if aggType == "Timer" {
		log.Printf("creating a Timer")
		a = new(Timer)
	}
	if aggType == "MovingTimer" {
		log.Printf("creating a MovingTimer")
		a = new(MovingTimer)
	}
	msg, ok := a.Parse(params, query)
	if !ok {
		log.Printf("Couldn't initialize aggregator with params %v:%v", params, msg)
		return nil
	}
	return
}

func ParseParameters(query string) (aggregator string, parameters []string, ok bool) {
	fields := strings.Split(query, " ", -1)
	ok = len(fields) > 1
	return fields[0], fields[1:], ok
}

type Timer struct {
	expr1 string
	expr2 string
	query string
}

func (t *Timer) String() string {
	return t.query
}

func (t *Timer) Parse(params []string, query string) (msg string, ok bool) {
	if len(params) != 2 {
		return "Must have two fields", false
	}
	t.expr1 = params[0]
	t.expr2 = params[1]
	t.query = query
	return fmt.Sprintf("Timing between %v and %v", t.expr1, t.expr2), true
}

func (t *Timer) Push(data JSONData) interface{} {
	val1, ok1 := GetDeep(t.expr1, data)
	val2, ok2 := GetDeep(t.expr2, data)
	if ok1 && ok2 {
		return val2.(float64) - val1.(float64)
	}
	return -1.
}

type MovingTimer struct {
	expr1 string
	expr2 string
	query string
	windowSize int
	window list.List
	sum float64
}

func (t *MovingTimer) String() string {
	return t.query
}

func (t *MovingTimer) Parse(params []string, query string) (msg string, ok bool) {
	if len(params) != 3 {
		return "Must have three fields", false
	}
	t.expr1 = params[0]
	t.expr2 = params[1]
	
	wSize, err := strconv.Atoi(params[2])
	if err != nil || wSize <= 0 {
		return fmt.Sprintf("Couldn't parse %v as an unsigned int", params[2]), false
	}
	t.windowSize = wSize
	t.query = query

	return fmt.Sprintf("Timing between %v and %v over window size %d", t.expr1, t.expr2, t.windowSize), true
}

func (t *MovingTimer) Push(data JSONData) interface{} {
	val1, ok1 := GetDeep(t.expr1, data)
	val2, ok2 := GetDeep(t.expr2, data)
	if ok1 && ok2 {
		time := val2.(float64) - val1.(float64)
		t.window.PushFront(time)
		t.sum += time
		if t.window.Len() > t.windowSize {
			lastElem := t.window.Back()
			t.sum -= lastElem.Value.(float64)
			t.window.Remove(lastElem)
		}
		return t.sum / float64(t.window.Len())
	}
	return -1.
}

type MovingHistogram struct {
	expr string
	query string
	windowSize int
	window list.List
	hist map[interface{}]int
}

func (t *MovingHistogram) Parse(params []string, query string) (msg string, ok bool) {
	if len(params) != 2 {
		return "Must have two fields", false
	}
	t.expr = params[0]
	
	wSize, err := strconv.Atoi(params[1])
	if err != nil || wSize <= 0 {
		return fmt.Sprintf("Couldn't parse %v as an unsigned int", params[2]), false
	}
	t.windowSize = wSize
	t.query = query

	return fmt.Sprintf("Moving histogram on %v over window size %d", t.expr, t.windowSize), true
}

func (t *MovingHistogram) Push(data JSONData) interface{} {
	val, ok := GetDeep(t.expr, data)
	if ok {
		t.window.PushFront(val)
		currentCount, ok := t.hist[val]
		if !ok {
			t.hist[val] = 1
		}else {
			t.hist[val] = currentCount + 1
		}
		if t.window.Len() > t.windowSize {
			lastElem := t.window.Back()
			t.hist[lastElem.Value] -= 1
			t.window.Remove(lastElem)
		}
		return t.hist
	}
	return -1.
}
