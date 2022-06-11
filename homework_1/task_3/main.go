package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

type Game struct {
	itsInProcess bool
	num1         int
	num2         int
	result       int
	resultCh     chan string
	expression   string
}

var gm = new(Game)

type client chan<- string

type msge struct {
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
	messages = make(chan msge)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	gm.resultCh = make(chan string)

	go starGame()
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

func starGame() {
	gm.itsInProcess = true
	rand.Seed(time.Now().UnixNano())
	gm.num1 = rand.Intn(100)
	sNume1 := fmt.Sprintf("%d", gm.num1)
	gm.num2 = rand.Intn(100)
	sNume2 := fmt.Sprintf("%d", gm.num2)
	switch rand.Intn(4) {
	case 0:
		gm.result = gm.num1 + gm.num2
		gm.resultCh <- sNume1 + " + " + sNume2 + " = ?"
		gm.expression = sNume1 + " + " + sNume2 + " = ?"
	case 1:
		gm.result = gm.num1 - gm.num2
		gm.resultCh <- sNume1 + " - " + sNume2 + " = ?"
		gm.expression = sNume1 + " - " + sNume2 + " = ?"
	case 2:
		gm.result = gm.num1 / gm.num2
		gm.resultCh <- sNume1 + " / " + sNume2 + " = ?"
		gm.expression = sNume1 + " / " + sNume2 + " = ?"
	case 3:
		gm.result = gm.num1 * gm.num2
		gm.resultCh <- sNume1 + " * " + sNume2 + " = ?"
		gm.expression = sNume1 + " * " + sNume2 + " = ?"
	}
}

func handleConn(c net.Conn) {
	defer c.Close()
	ch := make(chan string)
	go clientWriter(c, ch)
	fmt.Fprintln(c, "Please, enter your Nickname")
	var nick string
	fmt.Fscanln(c, &nick)
	whoIs := who{addr: c.RemoteAddr().String(), nickname: nick}
	ch <- "You are: " + whoIs.nickname + ", your address: " + whoIs.addr
	messages <- msge{m: whoIs.nickname + " has arrived"}
	entering <- ch
	log.Println(whoIs.nickname + " has arrived")
	input := bufio.NewScanner(c)
	for input.Scan() {
		answ := input.Text()
		answInt, _ := strconv.Atoi(answ)
		if answInt == gm.result {
			ch <- "you win"
			messages <- msge{m: "win", author: ch}
			gm.itsInProcess = false
		} else {
			ch <- "wrong answer"
		}
	}
	leaving <- ch
	messages <- msge{m: whoIs.nickname + " has left"}
	log.Println(whoIs.nickname + " has left")
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-gm.resultCh:
			for cli := range clients {
				cli <- msg
			}
		case msges := <-messages:
			if msges.m == "win" {
				for cli := range clients {
					if cli != msges.author {
						cli <- "you lose"
					}
				}
				go starGame()
			}
		case cli := <-entering:
			clients[cli] = true
			if !gm.itsInProcess {
				go starGame()
			} else {
				cli <- gm.expression
			}
		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}
