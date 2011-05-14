package main

import (
	//"fmt"
	"http"
	"websocket"
)

func ServeHello(socket *websocket.Conn) {
	hello := []byte("hello, world\n")
	socket.Write(hello);
}

func main() {
	http.Handle("/hello", websocket.Handler(ServeHello));
	err := http.ListenAndServe(":8080", nil);
	if err != nil {
		panic("ListenAndServe: " + err.String())
	}
}