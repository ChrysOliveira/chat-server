package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"

	"github.com/ChrysOliveira/chat-server/utils"
)

func main() {
	utils.CallClear()

	conn := connectToServer()

	done := make(chan struct{})
	go keepAlive(conn, done, comportamento)

	<-done // espera a gorrotina terminar
	conn.Close()
}

func connectToServer() net.Conn {
	conn, err := net.Dial("tcp", "localhost:3000")
	log.Println("Connected!")

	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte("BOT\n"))

	return conn
}

func keepAlive(conn net.Conn, ch chan struct{}, f func(s string) string) {
	input := bufio.NewScanner(conn)

	for input.Scan() {
		msgServer := input.Text()

		re := regexp.MustCompile(`^@(\w+)\sdisse\sem\sprivado:\s(.+)$`)
		matches := re.FindStringSubmatch(msgServer)
		if len(matches) == 0 {
			log.Println(msgServer)
		} else {
			log.Println("[" + strings.Join(matches, ",") + "]")

			r := f(matches[2])
			resposta := fmt.Sprintf("\\msg @%v %v\n", matches[1], r)
			log.Println(resposta)
			conn.Write([]byte(resposta))
		}
	}
	log.Println("Server connection lost")
	ch <- struct{}{}
}

func comportamento(s string) string {
	runes := []rune(s)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
