package main

import (
	"strings"
	"strconv"
	"fmt"
	"rand"
	"log"
)

type Filter interface {
	Parse(query string) (ok bool, applicable bool, msg string)
	Predicate(line JSONData) bool
}

func PassesAllFilters(line JSONData, filters []Filter) bool {
	for _, filter := range filters {
		if !filter.Predicate(line) {
			return false
		}
	}
	return true
}

func ParseFilters(statements []string) (filters []Filter, ok bool) {
	for i, statement := range statements {
		filter, ok := ParseStatement(statement)
		if ok {
			filters[i] = filter
		} else {
			log.Printf("Couldn't parse %s", statement)
			break
		}
	}
	return
}

func ParseStatement(statement string) (filter Filter, ok bool) {
	filter, ok = NewSamplingFilter(statement)
	if ok {
		return
	}
	return
}

type SamplingFilter struct {
	rate float64
}

func NewSamplingFilter(query string) (f *SamplingFilter, ok bool) {
	f = new(SamplingFilter)
	ok, applicable, msg := f.Parse(query)
	if applicable {
		log.Print(msg)
	} else {
		ok = false
	}
	return
}

func (f *SamplingFilter) Parse(query string) (ok bool, applicable bool, msg string) {
	fields := strings.Split(query, " ", -1)
	if fields[0] == "SAMPLE" {
		applicable = true
		if len(fields) == 2 {
			rate, err := strconv.Atof64(fields[1])
			if err != nil {
				ok = false
				msg = "Cannot parse rate " + fields[1] + " as a float"
				return
			}
			f.rate = rate
			if f.rate > 1 || f.rate < 0 {
				ok = false
				f.rate = -1
				msg = "Rate must be between 0 and 1"
			}
			return true, true, fmt.Sprintf("Sampling with rate %d", f.rate)
		}
	}
	return true, false, "N/A"
}

func (f SamplingFilter) Predicate(line JSONData) bool {
	return rand.Float64() < f.rate
}
