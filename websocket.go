package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"bufio"
	"http"
	"websocket"
	//"time"
	"log"
	"json"
	"strings"
	//"reflect"
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

func ServeWS(socket *websocket.Conn) {
	file, err := os.Open("ranger_sample.json")
	if err != nil {
		log.Fatal("Failed to open file", err)
	}
	defer file.Close()

	lineReader, err := bufio.NewReaderSize(file, 1024*16)
	if err != nil {
		log.Fatal("Failed to create reader %s", err)
	}

	for {
		line, isPrefix, err := lineReader.ReadLine()	
		if isPrefix {
			log.Fatal("Line is larger than buffer, TODO")
		}
		if err != nil {
			break
		}

		var data JSONData
		
		err = json.Unmarshal(line, &data)
		if err != nil {
			log.Fatal("Failure to decode: %s", err)
		}
		
		uniqueRequestID, ok := GetDeep("unique_request_id", data)

		if !ok {
			log.Print("Missing unique request id")
		}

		dirtySession := "NA"
		dirtySessionBool, ok := GetDeep("extra.dirty_session", data)
		if !ok {
			log.Print("Missing dirty session")	
		} else {
			dirtySession = fmt.Sprintf("%t", dirtySessionBool)
		}

		socket.Write([]byte(fmt.Sprintf("%s: %s", uniqueRequestID, dirtySession)))
	}
	
	/*
	hello := []byte("hello, world\n")
	for {
		//timer := time.NewTimer(1000000000);
		
		//fmt.Printf("Timer expired: %s\n", <- timer.C)
		<- time.After(100000000);
		socket.Write(hello);	
	}
	*/
}

func ServePage(writer http.ResponseWriter, request *http.Request) {
	//file, err := os.Open("default.do");
	log.Println("Starting /")
	
	//defer file.Close();
	contents, err := ioutil.ReadFile("websocket.html")
	if err != nil {
		fmt.Printf("Error opening file");
		return
	}

	writer.Write(contents);

}

func main() {
	log.Println("Starting up")
	
	http.Handle("/", http.HandlerFunc(ServePage));
	http.Handle("/ws", websocket.Handler(ServeWS));

	err := http.ListenAndServe(":8080", nil);
	if err != nil {
		panic("ListenAndServe: " + err.String())
	}
}