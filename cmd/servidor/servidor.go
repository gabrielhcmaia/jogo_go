package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func main() {
	log.Println("Inciando servidor na porta :8000")

	listner, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	defer listner.Close()

	for {
		conn, err := listner.Accept()
		if err != nil {
			log.Printf("Erro ao aceitar conexão: %v\n", err)
			continue
		}

		go HandleConnection(conn)
	}
}
func HandleConnection(conn net.Conn) {
	log.Printf("Novo jogador conectado: %s", conn.RemoteAddr().String())

	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Jogador %s se desconectou", conn.RemoteAddr().String())
			return
		}
		cleanMessage := strings.TrimSpace(message)
		log.Printf("Mensagem recebida de %s: %s", conn.RemoteAddr(), cleanMessage)
		conn.Write([]byte("Servidor recebeu: "+cleanMessage))
	}
}
