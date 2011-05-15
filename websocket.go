package main

import (
	"fmt"
	//"os"
	"io/ioutil"
	"http"
	"websocket"
	"time"
	"log"
)

func ServeWS(socket *websocket.Conn) {
	hello := []byte("hello, world\n")
	for {
		//timer := time.NewTimer(1000000000);
		
		//fmt.Printf("Timer expired: %s\n", <- timer.C)
		<- time.After(100000000);
		socket.Write(hello);	
	}
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