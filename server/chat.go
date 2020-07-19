// chat is a simple TCP/IP chat server.
package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

type client struct {
	name string
	ch   chan<- string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			// Broadcast the message to all connected clients.
			for cli := range clients {
				select {
				case cli.ch <- msg:
				default:
				}
			}
		case newcli := <-entering:
			// Announce the current set of clients to the new arrival.
			for cli := range clients {
				msg := cli.name + " is present"
				select {
				case newcli.ch <- msg:
				default:
				}
			}
			clients[newcli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli.ch)
		}
	}
}

func handleConn(conn net.Conn) {
	activity := make(chan time.Time, 2) // times at which client interacts with server
	activity <- time.Now()
	go clientDisconnector(conn, activity, 1*time.Minute)

	name := getClientName(conn)
	activity <- time.Now()
	ch := make(chan string, 1) // outbound messages to client
	cli := client{name, ch}

	go clientWriter(conn, ch)
	messages <- name + " has arrived"
	entering <- cli

	in := bufio.NewScanner(conn)
	for in.Scan() {
		messages <- name + ": " + in.Text()
		activity <- time.Now()
	}

	leaving <- cli
	messages <- name + " has left"
	conn.Close()
}

func getClientName(conn net.Conn) string {
	fmt.Fprint(conn, "Name: ")
	in := bufio.NewScanner(conn)
	in.Scan()
	name := in.Text()
	if len(name) == 0 {
		name = conn.RemoteAddr().String() // default to IP address
	}
	return name
}

func clientDisconnector(conn net.Conn, activity <-chan time.Time, timeout time.Duration) {
	lastActivity := <-activity // there will always be at least one activity (connection)
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case lastActivity = <-activity:
			// Do nothing.
		case <-ticker.C:
			if time.Since(lastActivity) > timeout {
				conn.Close()
				ticker.Stop()
				return
			}
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
