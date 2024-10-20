package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/ChrysOliveira/chat-server/utils"
)

type chanClient chan<- string // canal de mensagem

type User struct {
	chn     chanClient
	apelido string
}

type ChangeNick struct {
	old string
	new string
}

type PrivateMessage struct {
	usrSrc string
	usrDst string
	msg    string
}

var (
	entering          = make(chan User)
	enteringBroadcast = make(chan User)
	enteringCommand   = make(chan User)
	enteringPrivate   = make(chan User)

	leaving          = make(chan User)
	leavingBroadcast = make(chan User)
	leavingCommand   = make(chan User)
	leavingPrivate   = make(chan User)

	messages        = make(chan string)
	commands        = make(chan string)
	changeNick      = make(chan string)
	privateMessages = make(chan PrivateMessage)

	registerChannels = []chan User{
		enteringBroadcast,
		enteringCommand,
		enteringPrivate,
	}

	unregisterChannels = []chan User{
		leavingBroadcast,
		leavingCommand,
		leavingPrivate,
	}
)

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

	go registerChannelsHandler()
	go broadcasterHandler()
	go commandsHandler()
	go privateHandler()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func registerChannelsHandler() {
	for {
		select {
		case usr := <-entering:
			for _, v := range registerChannels {
				v <- usr
			}

		case usr := <-leaving:
			for _, v := range unregisterChannels {
				v <- usr
			}

			close(usr.chn)
		}
	}
}

func broadcasterHandler() {
	clients := make(map[chanClient]bool) // todos os clientes conectados
	for {
		select {

		case msg := <-messages:
			// broadcast de mensagens. Envio para todos
			// a chave eh o channel do cliente
			for cli := range clients {
				cli <- msg
			}
			log.Println(msg)

		case usr := <-enteringBroadcast:
			clients[usr.chn] = true

		case usr := <-leavingBroadcast:
			delete(clients, usr.chn)
		}
	}
}

func commandsHandler() {
	clients := make(map[chanClient]bool)
	re := regexp.MustCompile(`^(\\\w+)(?:\s+(.*))?$`)

	for {
		select {
		case cmd := <-commands:
			matches := re.FindStringSubmatch(cmd)

			switch matches[1] {
			case "\\changenickname":
				changeNick <- matches[2]
			}

		case usr := <-enteringCommand:
			clients[usr.chn] = true

		case usr := <-leavingCommand:
			delete(clients, usr.chn)
		}
	}
}

func privateHandler() {
	clients := make(map[string]chanClient)

	for {
		select {
		case newNick := <-changeNick:
			nicks := strings.Split(newNick, ",")
			chn := clients[nicks[0]]
			delete(clients, nicks[0])
			clients[nicks[1]] = chn

		case privateMessage := <-privateMessages:
			fmt.Printf("Src: %v | Dst: %v | Msg: %v\n", privateMessage.usrSrc, privateMessage.usrDst, privateMessage.msg)
			chnSrc := clients[privateMessage.usrDst[1:]] // [1:] para remover o @ do nick do dst
			frase := fmt.Sprintf("@%v disse em privado: %v\n", privateMessage.usrSrc, privateMessage.msg)
			chnSrc <- frase
			log.Println(frase)

		case usr := <-enteringPrivate:
			clients[usr.apelido] = usr.chn

		case usr := <-leavingPrivate:
			delete(clients, usr.apelido)
		}
	}
}

func handleConn(conn net.Conn) {
	bapelido := make([]byte, 20)

	size, err := conn.Read(bapelido)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: criar um map para add um id unico ao apelido para casos de apelidos duplicados
	apelido := string(bapelido[:size-1]) // size-1 removes the "\n"
	ch := make(chan string)

	usr := User{ch, apelido}

	go clientWriter(conn, ch)

	ch <- fmt.Sprintf("Connected as %q", usr.apelido)
	messages <- fmt.Sprintf("Usuario @%v acabou de entrar", usr.apelido)

	entering <- usr

	input := bufio.NewScanner(conn)

	for input.Scan() {
		re := regexp.MustCompile(`^(\\\w+)(?:\s+(@\w+))?(?:\s+(.*))?$`)
		rawTxt := input.Text()

		matches := re.FindStringSubmatch(rawTxt)

		if len(matches) == 0 {
			log.Printf("ERROR: Invalid command format at command %q\n", rawTxt)
		} else {
			cmd := matches[1]
			privateUser := matches[2]
			msg := matches[3]

			switch cmd {
			case "\\msg":
				if privateUser != "" {
					privateMessages <- PrivateMessage{usr.apelido, privateUser, msg}
				} else {
					messages <- fmt.Sprintf("@%v disse: %v", usr.apelido, msg)
				}
			case "\\exit":
				leaving <- usr
				conn.Close()
			case "\\changenickname":
				if msg != "" {
					commands <- fmt.Sprintf("%v %v,%v", cmd, usr.apelido, msg)
					messages <- fmt.Sprintf("Usuario @%v agora eh @%v", usr.apelido, msg)
					usr.apelido = msg
				} else {
					log.Println("Invalid nickname change command")
				}
			default:
				messages <- fmt.Sprintf("DEU RUIM: comando %v | msg: %v | matches: %v \n", cmd, msg, "["+strings.Join(matches, ",")+"]")
			}
		}
	}

	messages <- fmt.Sprintf("Usuario @%v saiu", usr.apelido)
}

func clientWriter(conn net.Conn, ch <-chan string) {
	for msg := range ch {
		fmt.Fprintln(conn, msg)
	}
}
