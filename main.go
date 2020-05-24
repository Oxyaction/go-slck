package main

import (
	"fmt"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":8081")
	if err != nil {
		panic(err)
	}

	h := NewHub()
	go h.run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
		}

		c := client{
			conn:       conn,
			outbound:   h.commands,
			register:   h.registrations,
			deregister: h.deregistrations,
		}

		go c.read()
	}
}
