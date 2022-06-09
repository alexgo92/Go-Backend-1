package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

type client chan<- string

type msg struct {
	m      string
	author client
}

type who struct {
	addr     string
	nickname string
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan msg)
)

func main() {
	conn, err := net.Listen("tcp", ":8002")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go broadcaster()

	for {
		l, err := conn.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handlerConn(l)

	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msge := <-messages:
			for cli := range clients {
				if msge.author != cli {
					cli <- msge.m
				}
			}
		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func handlerConn(l net.Conn) {
	ch := make(chan string)

	go clientWriter(l, ch)

	fmt.Fprintln(l, "Please, enter your Nickname")
	var nick string
	fmt.Fscanln(l, &nick)
	whoIs := who{addr: l.RemoteAddr().String(), nickname: nick}

	ch <- "You are: " + whoIs.nickname + ", your address: " + whoIs.addr
	messages <- msg{m: whoIs.nickname + " has arrived"}
	entering <- ch
	log.Println(whoIs.nickname + " has arrived")

	input := bufio.NewScanner(l)
	for input.Scan() {
		messages <- msg{m: whoIs.nickname + ": " + input.Text(), author: ch}
	}

	leaving <- ch
	messages <- msg{m: whoIs.nickname + " has left"}
	log.Println(whoIs.nickname + " has left")
	l.Close()
}

func clientWriter(l net.Conn, ch <-chan string) {
	for msge := range ch {
		fmt.Fprintln(l, msge)
	}
}
