package main

import (
	"os"
	"fmt"
	"container/list"
	"log"
)

type Window interface {
	Expression
	Push(element interface{}, wSize int) (err os.Error)
	Len() int
	RegisterPushCallback(cb windowCallback)
	GetPushCallback() windowCallback
	RegisterPopCallback(cb windowCallback)
	GetPopCallback() (cb windowCallback)
}

type windowCallback func(val interface{}) (err os.Error)

type RollingWindow struct {
	expr Expression
	window list.List
	windowSize Expression
	pushCb windowCallback
	popCb windowCallback
}
var _ Window = new(RollingWindow)

func (rw *RollingWindow) RegisterPushCallback(cb windowCallback) {
	rw.pushCb = cb
}

func (rw *RollingWindow) GetPushCallback() (cb windowCallback) {
	return rw.pushCb
}

func (rw *RollingWindow) RegisterPopCallback(cb windowCallback) {
	rw.popCb = cb
}

func (rw *RollingWindow) GetPopCallback() (cb windowCallback) {
	return rw.popCb
}

func (rw *RollingWindow) Len() int {
	return rw.window.Len()
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
		log.Printf("pushing %v", value)
		err = rw.Push(value, wSize.(int))
	}
	return rw.window.Front(), err
}

func (rw *RollingWindow) Push(element interface{}, wSize int) (err os.Error) {
	rw.window.PushFront(element)
	if rw.pushCb != nil {
		err = rw.pushCb(element)
	}
	if err != nil {
		return
	}
	if rw.window.Len() > wSize {
		lastElem := rw.window.Back()
		rw.window.Remove(lastElem)
		if rw.popCb != nil {
			err = rw.popCb(element)
		}
	}
	return
}

type WindowAve struct {
	window Window
	sum float64
}

func (wa *WindowAve) Setup(fname string, args []Expression) (err os.Error) {
	if len(args) != 1 {
		return fmt.Errorf("WindowAve expects a single Window argument.")
	}
	window, ok := args[0].(Window);
	if !ok {
		return fmt.Errorf("WindowAve expects a single Window argument.")
	}
	window.RegisterPushCallback(window.GetPushCallback())
	window.RegisterPopCallback(window.GetPopCallback())
	wa.window = window
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
