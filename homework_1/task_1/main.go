package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type client chan<- string

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
	go messageServer()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func handleConn(c net.Conn) {
	defer c.Close()
	ch := make(chan string)
	go clientWriter(c, ch)
	entering <- ch
	var err error
	for err == nil {
		time.Sleep(2 * time.Second)
		_, err = io.WriteString(c, time.Now().Format("15:04:05\n\r"))
	}
	leaving <- ch
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func messageServer() {
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		messages <- input.Text()
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msge := <-messages:
			for cli := range clients {
				cli <- msge
			}
		case cli := <-entering:
			clients[cli] = true
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
