package main

import (
	"strings"
	"os"
	"fmt"
)

func ParseFunction(statement string) (fname string, args []string, err os.Error) {
	fields := strings.Split(statement, "(", 2)
	fname = fields[0]
	argsStr := fields[1]
	if !strings.HasSuffix(argsStr, ")") {
		return "", nil, fmt.Errorf("statement must end in ')': %s", statement)
	}
	args = strings.Split(argsStr[:len(argsStr) - 1], ",", -1)
	err = nil
	return
}

type Function interface {
	Parse(args []string) (err os.Error)
}

/*
type Function struct {
	statement string
}

func (f *Function) String() string {
	return t.statement
}
**/

func ParseFilter(statement string) (f Filter, err os.Error) {
	fname, args, err := ParseFunction(statement)
	if err != nil {
		return nil, err
	}

	switch fname {
		case "RandomSample": 
			f = new(RandomSample)	
		default:
			return nil, fmt.Errorf("Unrecognized function name '%s', fname")
	}
	err = f.Parse(args)
	return
}

