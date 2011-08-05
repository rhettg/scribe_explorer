package main

import (
	"os"
	"fmt"
	"strconv"
	"log"
	"regexp"
)

func ParseString(statement string) (fname string, args []string, err os.Error) {
	expressionReString := `[A-Z|a-z|0-9]+\((.+)\)` //|([A-Z|a-z|0-9|\.]+)`
	expressionRe, err := regexp.Compile(expressionReString)
	if !expressionRe.MatchString(statement) {
		return "", []string{}, fmt.Errorf("\"%v\" is not an Expression", statement)
	}
	fnameRe, err := regexp.Compile(`([A-Z|a-z|0-9]+)\((.+)\)$`)
	fnameMatches := fnameRe.FindStringSubmatch(statement)
	fname = fnameMatches[1]
	argsStr := fnameMatches[2]

	// Scan over the arguments text, keeping track of the level of parentheses
	// nesting. If we reach a comma at the top-level, end the currentWord
	// and add it to the list of arguments.
	parenLevel := 0
	currentWord := []int{}
	for _, c := range argsStr {
		if c == '(' {
			parenLevel++
		} else if c == ')' {
			parenLevel--
		}

		if parenLevel == 0 && c == ',' {
			args = append(args, string(currentWord))
			currentWord = []int{}
		} else {
			currentWord = append(currentWord, c)
		}
	}
	// Don't forget to add the last word.
	args = append(args, string(currentWord))

	if parenLevel != 0 {
		return "", []string{}, fmt.Errorf("Unbalanced parentheses in \"%v\"", argsStr)
	}
	return fname, args, nil
}

type Expression interface {
	Setup(args []Expression) (err os.Error)
	Evaluate(data JSONData) (result interface{}, err os.Error)
	String() string
}

type Function struct {
	args []Expression
}

type Literal struct {
	value interface{}
}

func (l *Literal) Evaluate(data JSONData) (result interface{}, err os.Error) {
	return l.value, nil
}

func (l *Literal) Setup(args []Expression) (err os.Error) {
	// this ain't right
	return nil
}

func (l *Literal) String() string {
	return fmt.Sprintf("%v", l.value)
}

func ParseLiteral(literal string) (l *Literal, err os.Error) {
	l = new(Literal)
	if i, err := strconv.Atoi(literal); err == nil {
		l.value = i
		return l, nil
	} else if f, err := strconv.Atof64(literal); err == nil {
		l.value = f
		return l, nil
	} else if literal[0] == '"' && literal[len(literal)-1] == '"' {
		l.value = literal[1 : len(literal)-1]
		return l, nil
	}
	return nil, fmt.Errorf("Couldn't parse %s as a literal", literal)
}


func Parse(statement string) (expr Expression, err os.Error) {
	// First try to parse literals
	if expr, err = ParseLiteral(statement); err == nil {
		return
	}

	// Base case: statement is a single expression (e.g. Foo(a,b))
	fname, args, err := ParseString(statement)

	// Then treat it as a GetDeep expression if it's not an expression
	if err != nil {
		expr, err := NewGetDeepExpression(statement)
		if err == nil {
			log.Printf("found a get deep expr: %v, args: %v", expr, args)
			return expr, err
		}
		log.Printf("couldn't parse get deep expr: ", err)
	}

	// Now start parsing the rest
	expressionArgs := []Expression{}
	for _, arg := range args {
		argExpr, err := Parse(arg)
		if err != nil {
			return nil, err
		}
		expressionArgs = append(expressionArgs, argExpr)
	}

	switch fname {
	case "RandomSample":
		expr = new(RandomSample)
	case "GetDeep":
		expr = new(GetDeepExpression)
	case "Subtract":
		expr = new(Subtract)
	case "RollingAverage":
		expr = new(RollingAverage)
	default:
		return nil, fmt.Errorf("Unrecognized function name '%s'", fname)
	}
	err = expr.Setup(expressionArgs)
	return
}
