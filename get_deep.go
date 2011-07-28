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
	for _, subKey := range(allKeys) {
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
					log.Println("can't interpret %v as an array index", subKey)
					return nil, false
				}
				defer func() {
					if e := recover(); e != nil {
						log.Println("Array index out of bound")
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
