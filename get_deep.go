package main

import (
	"log"
	"strings"
)

type JSONData interface{}

func GetDeep(key string, data JSONData) (interface{}, bool) {
	allKeys := strings.Split(key, ".", -1)
	
	dataStep := data
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
				log.Fatal("don't know how to handle this one yet")
				return nil, false
			default:
				log.Fatal("don't konw how to handle this type: %s", dataType)
				return nil, false
		}
	}
	
	return dataStep, true
}