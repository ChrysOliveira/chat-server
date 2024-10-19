package main

import (
	"bufio"
	"fmt"
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

	mustCopy(conn, os.Stdin)

	conn.Close()
	// TODO: o processo nao esta encerrando quando perdemos a conexao
	<-done // espera a gorrotina terminar
}

func connectToServer() net.Conn {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Nickname: ")
	apelido, _ := reader.ReadString('\n')

	conn, err := net.Dial("tcp", "localhost:3000")
	log.Println("Connected!")

	if err != nil {
		log.Fatal(err)
	}

	conn.Write([]byte(apelido))

	return conn
}

func keepAlive(conn net.Conn, ch chan struct{}) {
	io.Copy(os.Stdout, conn)
	log.Println("Server connection lost")
	ch <- struct{}{}
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}
