package main

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/ChrysOliveira/chat-server/utils"
)

func main() {
	utils.CallClear()

	conn := connectToServer()

	done := make(chan struct{})
	go keepAlive(conn, done)
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

func keepAlive(conn net.Conn, ch chan struct{}) {
	io.Copy(os.Stdout, conn)
	log.Println("Server connection lost")
	ch <- struct{}{}
}
