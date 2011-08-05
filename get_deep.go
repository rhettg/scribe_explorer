package main

import (
	"log"
	"strings"
	"strconv"
)

type JSONData interface{}

func GetDeep(key string, data JSONData) (dataStep interface{}, ok bool) {
	allKeys := strings.Split(key, ".", -1)
	dataStep = data
	for _, subKey := range allKeys {
		// Check we have something sane we can use
		switch dataType := dataStep.(type) {
		case map[string]interface{}:
			value, ok := dataStep.(map[string]interface{})[subKey]
			if !ok {
				return nil, false
			}
			dataStep = value
			continue

		case []interface{}:
			arrayIndex, err := strconv.Atoi(subKey)
			if err != nil {
				return nil, false
			}
			defer func() {
				if e := recover(); e != nil {
					dataStep = nil
					ok = false
				}
			}()
			dataStep = dataStep.([]interface{})[arrayIndex]
			continue
		default:
			log.Println("don't know how to handle this type: %T", dataStep)
			return nil, false
		}
	}
	return dataStep, true
}


/*
 * GetDeepExpr
type GetDeepExpr struct {
	expr string
}

func (f *GetDeepExpr) Parse(args []string) (err os.Error) {
	if len(args) != 1 {
		return fmt.Errorf("GetDeepValue expects 1 argument, the GetDeep expression. Instead got %v", args)
	}
	f.expr = args[0]
}

func (f *GetDeepExpr) Evaluate() (data JSONData) {
	return GetDeep(f.expr, line)
}

func (f *GetDeepExpr) String() {
	return f.expr
}
*/
