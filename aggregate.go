package main

import (
	"strconv"
	"os"
	"container/list"
	"fmt"
)


/*
 * GetDeep(string) or string -> interface{}
 *
 * Performs a GetDeep() lookup on the JSONData. Returns
 */
type GetDeepExpression struct {
	expr Expression
}

func NewGetDeepExpression(expr string) (gd *GetDeepExpression, err os.Error) {
	gd = new(GetDeepExpression)
	exprLiteral, err := ParseLiteral(strconv.Quote(expr))
	if err != nil {
		return nil, err	
	}
	err = gd.Setup([]Expression{exprLiteral})
	return
}

func (gd *GetDeepExpression) Setup(args []Expression) (err os.Error) {
	if len(args) != 1 {
		return fmt.Errorf("GetDeep expects one argument, a string GetDeep expression")
	}
	gd.expr = args[0]
	return nil
}

func (gd *GetDeepExpression) Evaluate(data JSONData) (result interface{}, err os.Error) {
	key, err := gd.expr.Evaluate(data)
	if err != nil{
		return nil, err
	}
	if key, ok := key.(string); key == "" || !ok {
		return nil, fmt.Errorf("Expected non-empty string. Was type %T \"%v\"", key, key)
	}
	result, _ = GetDeep(key.(string), data)
	return
}

func (gd *GetDeepExpression) String() string {
	return gd.expr.String()
}


/*
 * Subtract(expr1, expr2 float64) -> float64
 */

type Subtract struct {
	expr1 Expression
	expr2 Expression
}

func (s *Subtract) Setup(args []Expression) (err os.Error) {
	if len(args) != 2 {
		return fmt.Errorf("Subtract expects two arguments, expressions that can be evaluated to numeric types")
	}
	s.expr1, s.expr2 = args[0], args[1]
	return nil
}

func (s *Subtract) Evaluate(data JSONData) (result interface{}, err os.Error) {
	val1, err1 := s.expr1.Evaluate(data)
	val2, err2 := s.expr2.Evaluate(data)
	if err1 != nil {
		return nil, fmt.Errorf("Expression 1 could not be evaluated, %v", err2)
	}
	if err2 != nil {
		return nil, fmt.Errorf("Expression 1 could not be evaluated, %v", err2)
	}
	val1, ok1 := val1.(float64)
	val2, ok2 := val2.(float64)
	if !ok1 {
		return nil, fmt.Errorf("Subtract expects a float64, Expression 1 was type %T, val %v", val1, val1)
	}
	if !ok2 {
		return nil, fmt.Errorf("Subtract expects a float64, Expression 2 was type %T, val %v", val2, val1)
	}

	return val1.(float64) - val2.(float64), nil
}

func (s *Subtract) String() string {
	return fmt.Sprintf("Subtract(%v,%v)", s.expr1, s.expr2)
}

/*
 * RollingAverage(x float64, windowSize int) -> float64
 */
type RollingAverage struct {
	expr Expression
	windowSize Expression
	window list.List
	sum float64
}

func (ra *RollingAverage) String() string {
	return fmt.Sprintf("RollingAverage(%v,%v)", ra.expr, ra.windowSize)
}

func (ra *RollingAverage) Setup(args []Expression) (err os.Error) {
	if len(args) != 2 {
		return fmt.Errorf("RollingAverage must have 2 args, a float64 value, and a positive int window size. Got %v", args)
	}
	ra.expr = args[0]
	ra.windowSize = args[1]
	
	return nil
}

func (ra *RollingAverage) Evaluate(data JSONData) (result interface{}, err os.Error) {
	value, err := ra.expr.Evaluate(data)
	if err != nil {
		return nil, err
	}
	if value, ok := value.(float64); !ok {
		return nil, fmt.Errorf("RollingAverage expects a float64, got a %T, %v", value, value)
	}
	wSize, err := ra.windowSize.Evaluate(data)
	if err != nil {
		return nil, err
	}
	if wSize, ok := wSize.(int); !ok {
		return nil, fmt.Errorf("RollingAverage expects an int window size. Got a %T, %v", wSize, wSize)
	}
	ave := ra.Push(value.(float64), wSize.(int))
	return ave, nil
}

func (ra *RollingAverage) Push(val float64, windowSize int) float64 {
	ra.window.PushFront(val)
	ra.sum += val 
	if ra.window.Len() > windowSize {
		lastElem := ra.window.Back()
		ra.sum -= lastElem.Value.(float64)
		ra.window.Remove(lastElem)
	}
	return ra.sum / float64(ra.window.Len())
}
