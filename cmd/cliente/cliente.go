package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var currentBoard [3][3]string
var mySymbol string
var currentTurn string

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

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func DrawBoard() {
	clearScreen()
	fmt.Println("===JOGO DA VELHA===")
	fmt.Println()
	for r := 0; r < 3; r++ {
		fmt.Print(" ")
		for c := 0; c < 3; c++ {
			fmt.Print(currentBoard[r][c] + " ")
		}
		fmt.Println()
	}
	fmt.Println()
}

func readServerMessages(conn net.Conn) {
	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("O servidor desconectou")
			os.Exit(1)
		}
		message = strings.TrimSpace(message)

		if strings.HasPrefix(message, "MSG:") {
			fmt.Println(strings.TrimPrefix(message, "MSG:"))
		} else if strings.HasPrefix(message, "ERROR:") {
			fmt.Println("ERRO: " + strings.TrimPrefix(message, "ERROR:") + "!!!")
		} else if strings.HasPrefix(message, "BOARD:") {
			boardData := strings.TrimPrefix(message, "BOARD:")
			rows := strings.Split(boardData, ";")
			for r, rowData := range rows {
				for c, char := range rowData {
					currentBoard[r][c] = string(char)
				}
			}
			DrawBoard()
		} else if strings.HasPrefix(message, "SYMBOL:") {
			mySymbol = strings.TrimPrefix(message, "SYMBOL:")
			fmt.Printf("== Você é o Jogador %s ==\n", mySymbol)
		} else if strings.HasPrefix(message, "TURN:") {
			currentTurn = strings.TrimPrefix(message, "TURN:")
			if currentTurn == mySymbol {
				fmt.Println(">> É sua vez, digite (linha,coluna):")
			} else {
				fmt.Printf(">> Vez do jogador: %s, aguarde ...\n", currentTurn)
			}
		}
	}
}
