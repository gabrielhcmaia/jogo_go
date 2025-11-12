package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

type Game struct {
	board         [3][3]string
	players       [2]net.Conn
	playerCount   int
	currentPlayer int
}

type PlayerAction struct {
	playerIndex int
	message     string
	conn        net.Conn
	actionType  string
}

var game Game
var actions = make(chan PlayerAction)

func main() {
	log.Println("Iniciando servidor (com Canais) na porta :8000")

	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	resetGame()
	go gameManager()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Erro ao aceitar conexão: %v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func gameManager() {
	for action := range actions {
		switch action.actionType {
		case "CONNECT":
			handlePlayerConnect(action.conn)
		case "PLAY":
			processarJogada(action.playerIndex, action.message)
		case "DISCONNECT":
			handlePlayerDisconnect(action.playerIndex)
		}
	}
}

func handlePlayerConnect(conn net.Conn) {
	var playerIndex int
	var playerSymbol string

	if game.playerCount == 0 {
		playerIndex = 0
		playerSymbol = "X"
		game.players[playerIndex] = conn
		game.playerCount++

		log.Printf("Jogador 1 (%s) conectado: %s", conn.RemoteAddr().String(), playerSymbol)
		fmt.Fprintln(conn, "SYMBOL:X")
		fmt.Fprintln(conn, "MSG:BEMVINDO JOGADOR 1. VOCÊ É O 'X'.")
		fmt.Fprintln(conn, "MSG:Aguardando Jogador 2...")

	} else if game.playerCount == 1 {
		playerIndex = 1
		playerSymbol = "O"
		game.players[playerIndex] = conn
		game.playerCount++

		log.Printf("Jogador 2 (%s) conectado: %s", conn.RemoteAddr().String(), playerSymbol)
		fmt.Fprintln(conn, "SYMBOL:O")
		fmt.Fprintln(conn, "MSG:BEMVINDO JOGADOR 2. VOCÊ É O 'O'.")

		broadcast("MSG:O JOGO COMEÇOU.")
		broadcast(getBoardString())
		broadcast(fmt.Sprintf("TURN:%s", "X"))

	} else {
		log.Printf("Conexão recusada (jogo cheio): %s", conn.RemoteAddr().String())
		fmt.Fprintln(conn, "ERROR:DESCULPE, O JOGO JÁ ESTÁ CHEIO.")
		conn.Close()
		return
	}

	go listenForPlayerMessages(conn, playerIndex)
}

func handleConnection(conn net.Conn) {
	actions <- PlayerAction{
		conn:       conn,
		actionType: "CONNECT",
	}
}

func listenForPlayerMessages(conn net.Conn, playerIndex int) {
	reader := bufio.NewReader(conn)
	playerSymbol := "X"
	if playerIndex == 1 {
		playerSymbol = "O"
	}

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Jogador %s (%s) desconectou (leitura).", conn.RemoteAddr(), playerSymbol)
			actions <- PlayerAction{
				playerIndex: playerIndex,
				actionType:  "DISCONNECT",
			}
			return
		}

		cleanMessage := strings.TrimSpace(message)
		log.Printf("Mensagem de (%s): %s", playerSymbol, cleanMessage)

		actions <- PlayerAction{
			playerIndex: playerIndex,
			message:     cleanMessage,
			actionType:  "PLAY",
		}
	}
}

func handlePlayerDisconnect(playerIndex int) {
	playerSymbol := "X"
	if playerIndex == 1 {
		playerSymbol = "O"
	}

	log.Printf("Jogador %s (%d) desconectou.", playerSymbol, playerIndex)
	opponentIndex := (playerIndex + 1) % 2
	if game.players[opponentIndex] != nil {
		fmt.Fprintln(game.players[opponentIndex], fmt.Sprintf("MSG:JOGADOR %s DESCONECTOU.", playerSymbol))
	}
	resetGame()
}

func processarJogada(playerIndex int, message string) {
	conn := game.players[playerIndex]
	playerSymbol := "X"
	if playerIndex == 1 {
		playerSymbol = "O"
	}

	if game.playerCount < 2 {
		fmt.Fprintln(conn, "ERROR:Esperando o outro jogador entrar.")
		return
	}
	if playerIndex != game.currentPlayer {
		fmt.Fprintln(conn, "ERROR:Não é a sua vez.")
		return
	}

	coords := strings.Split(message, ",")
	if len(coords) != 2 {
		fmt.Fprintln(conn, "ERROR:Formato inválido. Use: linha,coluna")
		return
	}

	row, errRow := strconv.Atoi(coords[0])
	col, errCol := strconv.Atoi(coords[1])

	if errRow != nil || errCol != nil || row < 0 || row > 2 || col < 0 || col > 2 {
		fmt.Fprintln(conn, "ERROR:Posição inválida. Use números de 0 a 2.")
		return
	}
	if game.board[row][col] != "" {
		fmt.Fprintln(conn, "ERROR:Posição já ocupada.")
		return
	}

	game.board[row][col] = playerSymbol
	broadcast(fmt.Sprintf("MSG:JOGADA: %s jogou em %d,%d", playerSymbol, row, col))
	broadcast(getBoardString())

	if checkWin(playerSymbol) {
		broadcast(fmt.Sprintf("MSG:FIM DE JOGO! JOGADOR %s VENCEU!", playerSymbol))
		broadcast(getBoardString())
		resetGame()
		return
	}

	isDraw := true
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			if game.board[r][c] == "" {
				isDraw = false
			}
		}
	}
	if isDraw {
		broadcast("MSG:FIM DE JOGO! EMPATE!")
		broadcast(getBoardString())
		resetGame()
		return
	}

	game.currentPlayer = (game.currentPlayer + 1) % 2
	nextPlayerSymbol := "X"
	if game.currentPlayer == 1 {
		nextPlayerSymbol = "O"
	}
	broadcast(fmt.Sprintf("TURN:%s", nextPlayerSymbol))
}

func broadcast(message string) {
	log.Printf("Broadcast: %s", message)
	for i := 0; i < game.playerCount; i++ {
		if game.players[i] != nil {
			fmt.Fprintln(game.players[i], message)
		}
	}
}

func getBoardString() string {
	var b strings.Builder
	b.WriteString("BOARD:")
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			spot := game.board[r][c]
			if spot == "" {
				spot = "."
			}
			b.WriteString(spot)
		}
		if r < 2 {
			b.WriteString(";")
		}
	}
	return b.String()
}

func checkWin(symbol string) bool {
	for i := 0; i < 3; i++ {
		if (game.board[i][0] == symbol && game.board[i][1] == symbol && game.board[i][2] == symbol) ||
			(game.board[0][i] == symbol && game.board[1][i] == symbol && game.board[2][i] == symbol) {
			return true
		}
	}
	if (game.board[0][0] == symbol && game.board[1][1] == symbol && game.board[2][2] == symbol) ||
		(game.board[0][2] == symbol && game.board[1][1] == symbol && game.board[2][0] == symbol) {
		return true
	}
	return false
}

func resetGame() {
	log.Println("Resetando o jogo...")
	game.board = [3][3]string{}
	game.players = [2]net.Conn{}
	game.playerCount = 0
	game.currentPlayer = 0
}
