package main

import (
	"log"
	"net/textproto"
	"io"
)

func CreateStream(connectString string) io.Reader {
	c, err := textproto.Dial("tcp4", connectString)
	if err != nil {
		log.Fatal("Failed to open", err)
	}

	id, err := c.Cmd("ranger")
	if err != nil {
		log.Fatal("Failed to send cmd", err)
	}
	log.Printf("Sent command", id)

	// Return the reader part
	return c.R
}