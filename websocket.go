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
)

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

		var genericData interface{}
		
		
		err = json.Unmarshal(line, &genericData)
		if err != nil {
			log.Fatal("Failure to decode: %s", err)
		}
		
		data := genericData.(map[string]interface{})
		uniqueRequestID, ok := data["unique_request_id"].(string)
		if ok {
			socket.Write([]byte(uniqueRequestID))
		} else {
			log.Print("Missing unique request id")
		}

		
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