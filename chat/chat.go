// chat is a simple TCP/IP chat server.
package main

import (
	"fmt"
	"log"
	"net"
)

// client represents a single client of the chat server.
// Has a name and a send-only channel for outbound messages.
type client struct {
	name string
	ch   chan<- string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string) // clients write to this
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

// handleConn handles a new TCP client connection.
// Starts goroutines to handle client reading, writing, and disconnect-on-idle.
func handleConn(conn net.Conn) {
	ch := make(chan string)
	go clientWriter(conn, ch)

}

func clientReader() {

}

// clientWriter forwards every string sent via channel ch onto
// the parameterized network connection conn.
func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func clientDisconnector() {

}
