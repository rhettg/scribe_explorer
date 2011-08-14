package main

import (
	"os"
	"fmt"
	"container/list"
	"log"
	"time"
)

type Window interface {
	Expression
	Push(element interface{}, wSize int) (err os.Error)
	Len() int
	SetListener(l WindowListener)
}

type windowCallback func(val interface{}) (err os.Error)

type WindowListener interface {
	Push(element interface{}) (err os.Error)
	Pop(element interface{}) (err os.Error)
}

type SingleWindowListenerStruct struct {
	window Window
}

type RollingWindow struct {
	expr       Expression
	windowList list.List
	windowSize Expression
	listener   WindowListener
}

var _ Window = new(RollingWindow)

func (rw *RollingWindow) Len() int {
	return rw.windowList.Len()
}

func (rw *RollingWindow) String() string {
	return fmt.Sprintf("RollingWindow(%v,%v)", rw.expr, rw.windowSize)
}

func (rw *RollingWindow) Setup(fname string, args []Expression) (err os.Error) {
	if len(args) != 2 {
		return fmt.Errorf("RollingWindow must have 2 args, the element and a positive int window size. Got %v", args)
	}
	rw.expr = args[0]
	rw.windowSize = args[1]

	return nil
}

func (rw *RollingWindow) SetListener(l WindowListener) {
	rw.listener = l
}

func (rw *RollingWindow) Evaluate(data JSONData) (result interface{}, err os.Error) {
	value, err := rw.expr.Evaluate(data)
	if err != nil {
		return nil, err
	}

	wSize, err := rw.windowSize.Evaluate(data)
	if err != nil {
		return nil, err
	}
	wSize, ok := wSize.(int)
	if !ok {
		return nil, fmt.Errorf("RollingWindow expects an int window size. Got a %T, %v", wSize, wSize)
	}
	if value != nil {
		log.Printf("pushing %v, wSize %v", value, wSize.(int))
		err = rw.Push(value, wSize.(int))
	}
	return rw.windowList.Front(), err
}

func (rw *RollingWindow) Push(element interface{}, wSize int) (err os.Error) {
	rw.windowList.PushFront(element)
	if rw.listener != nil {
		err = rw.listener.Push(element)
	}
	if err != nil {
		return
	}
	log.Printf("wSize %v", wSize)
	for rw.windowList.Len() > wSize {
		log.Printf("trimming window to %v", wSize)
		lastElem := rw.windowList.Back()
		rw.windowList.Remove(lastElem)
		if rw.listener != nil {
			err = rw.listener.Pop(lastElem.Value)
		}
	}
	return
}

type TimedWindow struct {
	expr         Expression
	windowList   list.List
	windowLength Expression
	listener     WindowListener
}

type timedWindowElement struct {
	value     interface{}
	timestamp int64
}

var _ Window = new(TimedWindow)

func (tw *TimedWindow) Len() int {
	return tw.windowList.Len()
}

func (tw *TimedWindow) String() string {
	return fmt.Sprintf("TimedWindow(%v,%v)", tw.expr, tw.windowLength)
}

func (tw *TimedWindow) Setup(fname string, args []Expression) (err os.Error) {
	if len(args) != 2 {
		return fmt.Errorf("RollingWindow must have 2 args, the element and a positive int window size. Got %v", args)
	}
	tw.expr = args[0]
	tw.windowLength = args[1]

	return nil
}

func (tw *TimedWindow) SetListener(l WindowListener) {
	tw.listener = l
}

func (tw *TimedWindow) Evaluate(data JSONData) (result interface{}, err os.Error) {
	value, err := tw.expr.Evaluate(data)
	if err != nil {
		return nil, err
	}

	wSize, err := tw.windowLength.Evaluate(data)
	if err != nil {
		return nil, err
	}
	wSize, ok := wSize.(int)
	if !ok {
		return nil, fmt.Errorf("RollingWindow expects an int (number of seconds) window size. Got a %T, %v", wSize, wSize)
	}
	if value != nil {
		err = tw.Push(value, wSize.(int))
	}
	return tw.windowList.Front(), err
}

func (tw *TimedWindow) Push(element interface{}, wSize int) (err os.Error) {
	now := time.Seconds()
	tw.windowList.PushFront(timedWindowElement{element, now})
	if tw.listener != nil {
		err = tw.listener.Push(element)
	}
	if err != nil {
		return
	}

	// Now trim off any elements that occured before the beginning of the window.
	windowStart := now - int64(wSize)
	for {
		backElem := tw.windowList.Back()
		if backElem == nil {
			return
		}
		backVal := backElem.Value.(timedWindowElement)
		if backVal.timestamp < windowStart {
			tw.windowList.Remove(backElem)
			if tw.listener != nil {
				err = tw.listener.Pop(backVal.value)
			}
		} else {
			return
		}
	}
	return
}

type WindowAve struct {
	window Window
	sum    float64
}

var _ WindowListener = new(WindowAve)

func (wa *WindowAve) Setup(fname string, args []Expression) (err os.Error) {
	if len(args) != 1 {
		return fmt.Errorf("WindowAve expects a single Window argument.")
	}
	window, ok := args[0].(Window)
	if !ok {
		return fmt.Errorf("WindowAve expects a single Window argument.")
	}
	wa.window = window
	wa.window.SetListener(wa)
	return
}

func (wa *WindowAve) Evaluate(data JSONData) (result interface{}, err os.Error) {
	wa.window.Evaluate(data)
	if wa.window.Len() == 0 {
		return 0., fmt.Errorf("Empty window")
	}
	return wa.sum / float64(wa.window.Len()), nil
}

func (wa *WindowAve) Push(val interface{}) (err os.Error) {
	log.Printf("Pushing %v on ave", val)
	if val, ok := val.(float64); !ok {
		return fmt.Errorf("Window expected a float64, got %v (%T)", val, val)
	}
	wa.sum += val.(float64)
	return nil
}

func (wa *WindowAve) Pop(val interface{}) (err os.Error) {
	if val, ok := val.(float64); !ok {
		return fmt.Errorf("Window expected a float64, got %v (%T)", val, val)
	}
	wa.sum -= val.(float64)
	return nil
}

func (wa *WindowAve) String() string {
	return fmt.Sprintf("WindowAve(%v)", wa.window)
}
