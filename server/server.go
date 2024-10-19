package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ChrysOliveira/chat-server/utils"
)

type client chan<- string // canal de mensagem

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func broadcaster() {
	clients := make(map[client]bool) // todos os clientes conectados
	for {
		select {

		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			// a chave eh o channel do cliente
			for cli := range clients {
				cli <- msg
			}

		case cli := <-entering:
			clients[cli] = true

		case cli := <-leaving:
			delete(clients, cli)
			close(cli)
		}
	}
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}

func handleConn(conn net.Conn) {
	ch := make(chan string)

	go clientWriter(conn, ch)

	bapelido := make([]byte, 20)
	size, err := conn.Read(bapelido)
	if err != nil {
		log.Fatal(err)
	}
	apelido := string(bapelido[:size-1]) // size-1 removes the "\n"

	input := bufio.NewScanner(conn)

	ch <- fmt.Sprintf("Connected as %q", apelido)
	messages <- apelido + " chegou!"
	entering <- ch

	for input.Scan() {
		messages <- apelido + ":" + input.Text()
	}

	leaving <- ch
	messages <- apelido + " se foi "
	conn.Close()
}

func main() {
	fmt.Println("Iniciando servidor...")

	listener, err := net.Listen("tcp", "localhost:3000")
	if err != nil {
		log.Fatal(err)
	}

	// TODO: melhorar
	await := time.After(time.Second)
	<-await
	utils.CallClear()
	fmt.Println("Servidor iniciado!")

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
