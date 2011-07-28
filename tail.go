package main

import (
	"log"
	"net"
	"io"
)

func CreateStream(connectString string) io.Reader {
	c, err := net.Dial("tcp4", connectString)
	if err != nil {
		log.Fatal("Failed to open", err)
	}

	id, err := c.Write([]uint8("ranger\n"))
	if err != nil {
		log.Fatal("Failed to send cmd", err)
	}
	log.Printf("Sent command", id)

	// Return the reader part
	return c
}