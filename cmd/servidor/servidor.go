package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Game struct {
	board         [3][3]string
	players       [2]net.Conn
	playerCount   int
	currentPlayer int
	mu            sync.Mutex
}

var game Game

func main() {
	log.Println("Inciando servidor na porta :8000")

	listner, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}

	defer listner.Close()
	resetGame()

	game.currentPlayer = 0

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

	game.mu.Lock()

	var playerSymbol string
	var playerIndex int

	if game.playerCount == 0 {
		playerIndex = 0
		playerSymbol = "X"
		game.players[playerIndex] = conn
		game.playerCount++

		log.Printf("Jogador 1 (X) conectado: %s", conn.RemoteAddr().String())
		fmt.Fprintln(conn, "BEMVINDO JOGADOR 1. VOCÊ É O 'X'")
		fmt.Fprintln(conn, "Aguardando Jogador 2...")
	} else if game.playerCount == 1 {
		playerIndex = 1
		playerSymbol = "O"
		game.players[playerIndex] = conn
		game.playerCount++

		log.Printf("Jogador 2 (O) conectado: %s", conn.RemoteAddr().String())
		fmt.Fprintln(conn, "BEMVINDO JOGADOR 2. VOCÊ É O 'O'")
		fmt.Fprintln(conn, "O JOGO VAI COMEÇAR.")
		broadcast(getBoardString())
		broadcast("Turno do jogador X")

	} else {
		log.Printf("Conexão recusada (jogo cheio): %s", conn.RemoteAddr().String())
		fmt.Fprintln(conn, "Jogo Cheio")
		conn.Close()
		game.mu.Unlock()
		return
	}
	game.mu.Unlock()

	log.Printf("Novo jogador conectado: %s", conn.RemoteAddr().String())

	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Jogador %s (%s) desconectou.", conn.RemoteAddr(), playerSymbol)
			game.mu.Lock()
			opponentIndex := (playerIndex + 1) % 2
			if game.players[opponentIndex] != nil {
				fmt.Fprintf(game.players[opponentIndex], "%s", fmt.Sprintf("JOGADOR %s DESCONECTOU.", playerSymbol))
			}
			resetGame()
			game.mu.Unlock()
			return
		}

		cleanMessage := strings.TrimSpace(message)
		log.Printf("Mensagem de (%s): %s", playerSymbol, cleanMessage)
		processarJogada(playerIndex, cleanMessage)
	}
}

func processarJogada(playerIndex int, message string) {
	game.mu.Lock()
	defer game.mu.Unlock()

	conn := game.players[playerIndex]
	playerSymbol := "X"
	if playerIndex == 1 {
		playerSymbol = "O"
	}

	if game.playerCount < 2 {
		fmt.Fprintln(conn, "Erro: esperando outro jogador")
		return
	}

	if playerIndex != game.currentPlayer {
		fmt.Println(conn, "Erro: não é sua vez")
		return
	}

	coords := strings.Split(message, ",")
	if len(coords) != 2 {
		fmt.Fprintln(conn, "ERRO: Formato inválido. Use: linha,coluna (ex: 0,2)")
		return
	}

	row, errRow := strconv.Atoi(coords[0])
	col, errCol := strconv.Atoi(coords[1])

	if errRow != nil || errCol != nil || row < 0 || row > 2 || col < 0 || col > 2 {
		fmt.Fprintln(conn, "ERRO: Posição inválida. Use números de 0 a 2.")
		return
	}

	if game.board[row][col] != "" {
		fmt.Fprintln(conn, "ERRO: Posição já ocupada.")
		return
	}
	game.board[row][col] = playerSymbol
	broadcast(fmt.Sprintf("JOGADA: %s jogou em %d,%d", playerSymbol, row, col))
	broadcast(getBoardString())

	if checkWin(playerSymbol) {
		broadcast(fmt.Sprintf("FIM DE JOGO! JOGADOR %s VENCEU!", playerSymbol))
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
		broadcast("FIM DE JOGO! EMPATE!")
		resetGame()
		return
	}

	game.currentPlayer = (game.currentPlayer + 1) % 2 // Alterna 0 -> 1 -> 0
	nextPlayerSymbol := "X"
	if game.currentPlayer == 1 {
		nextPlayerSymbol = "O"
	}
	broadcast(fmt.Sprintf("TURNO DO JOGADOR: %s", nextPlayerSymbol))

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
	b.WriteString("\nTABULEIRO:\n")
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			spot := game.board[r][c]
			if spot == "" {
				spot = "."
			}
			b.WriteString(spot + " ")
		}
		b.WriteString("\n")
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
