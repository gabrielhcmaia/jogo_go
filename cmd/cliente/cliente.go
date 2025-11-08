package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	log.Println("Conectando ao servidor na porta: localhost:8000")

	conn, err := net.Dial("tcp", "localhost:8000")

	if err != nil {
		log.Fatal(("Não foi possível se conectar "), err)
	}

	defer conn.Close()

	log.Println("Conectado!")
	go readServerMessages(conn)

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		text := scanner.Text()

		_, err := fmt.Fprintf(conn, "%s\n", text)
		if err != nil {
			log.Println("Erro ao enviar mensagem: ", err)
			break
		}
	}
}

func readServerMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("O servidor desconectou")
			os.Exit(1)
		}
		fmt.Print(message)
	}
}
