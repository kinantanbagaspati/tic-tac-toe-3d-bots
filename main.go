package main

import "fmt"

var length int = 4
var width int = 4
var height int = 4
var winLength int = 4
var board [4][4][4]byte

func initBoard() {
	for i := 0; i < length; i++ {
		for j := 0; j < width; j++ {
			for k := 0; k < height; k++ {
				board[i][j][k] = '|'
			}
		}
	}
}

func printBoard() {
	toPrint := make([][]byte, length+width+height-2)
	for i := range toPrint {
		toPrint[i] = make([]byte, length*width)
		for j := range toPrint[i] {
			toPrint[i][j] = ' '
		}
	}

	for i := 0; i < length; i++ {
		for j := 0; j < width; j++ {
			for k := 0; k < height; k++ {
				toPrint[i + width-j + height-k - 2][i*width + j] = board[i][j][k]
			}
		}
	}

	for i := range toPrint {
		fmt.Println(string(toPrint[i]))
	}
}

// Move should be A3, B1, ..., L12, etc.
// player should be 'x' or 'o'
func move(move string, player byte, board *[4][4][4]byte) (int, int, int) {
	if len(move) < 2 {
		return -1, -1, -1
	}
	// Get column
	col := move[0] - 'A'
	if col < 0 || col >= byte(length) {
		return -1, -1, -1
	}
	// Get row
	row := 0
	for i := 1; i < len(move); i++ {
		if move[i] < '0' || move[i] > '9' {
			return -1, -1, -1
		}
		row = row*10 + int(move[i]-'0')
	}
	row--
	if row < 0 || row >= width {
		return -1, -1, -1
	}
	// Find the lowest empty spot in the column
	for h := 0; h < height; h++ {
		if board[col][row][h] == '|' {
			board[col][row][h] = player
			return int(col), row, h
		}
	}
	return -1, -1, -1
}

func isValideCoordinate(x, y, z int) bool {
	return x >= 0 && x < length && y >= 0 && y < width && z >= 0 && z < height
}

func getLine(board *[4][4][4]byte, start [3]int, direction [3]int) [4]byte {
	var line [4]byte
	for i := 0; i < winLength; i++ {
		x := start[0] + i*direction[0]
		y := start[1] + i*direction[1]
		z := start[2] + i*direction[2]
		if !isValideCoordinate(x, y, z) {
			return [4]byte{'|', '|', '|', '|'}
		}
		line[i] = board[x][y][z]
	}
	return line
}

// Check all possible winning lines in 3D Tic-Tac-Toe
// Return 'x' if player X wins, 'o' if player O wins, or '|' if no winner
func checkWin(board *[4][4][4]byte) byte {
	directions := [][3]int{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, // 1D
		{1, 1, 0}, {1, -1, 0}, {1, 0, 1}, {1, 0, -1}, {0, 1, 1}, {0, 1, -1}, // 2D diagonals
		{1, 1, 1}, {1, -1, -1}, {1, 1, -1}, {1, -1, 1}, // 3D diagonals
	}
	for i := 0; i < length; i++ {
		for j := 0; j < width; j++ {
			for k := 0; k < height; k++ {
				// Check all 13 directions from each cell
				for _, dir := range directions {
					if !isValideCoordinate(i+ (winLength-1)*dir[0], j+(winLength-1)*dir[1], k+(winLength-1)*dir[2]) {
						continue
					}
					line := getLine(board, [3]int{i, j, k}, dir)
					if string(line[:]) == "xxxx" {
						for n := 0; n < winLength; n++ {
							board[i + n*dir[0]][j + n*dir[1]][k + n*dir[2]] = 'X'
						}
						return 'x'
					}
					if string(line[:]) == "oooo" {
						for n := 0; n < winLength; n++ {
							board[i + n*dir[0]][j + n*dir[1]][k + n*dir[2]] = 'O'
						}
						return 'o'
					}
				}
			}
		}
	}
	return '|'
}

func main() {
	initBoard()
	
	players := []byte{'x', 'o'}
	playerNames := []string{"Player X", "Player O"}
	currentPlayer := 0
	totalMoves := 0
	maxMoves := length * width * height
	
	fmt.Println("Welcome to 3D Tic-Tac-Toe!")
	fmt.Printf("Enter moves in format like A1, B2, etc. (A-%c, 1-%d)\n", 'A'+byte(length-1), width)
	fmt.Println()
	
	for totalMoves < maxMoves {
		printBoard()
		fmt.Printf("\n%s's turn (playing '%c'): ", playerNames[currentPlayer], players[currentPlayer])
		
		var moveInput string
		fmt.Scanln(&moveInput)
		
		x, y, z := move(moveInput, players[currentPlayer], &board)

		if x == -1 && y == -1 && z == -1 {
			fmt.Println("Invalid move! Try again.")
			continue
		}
		
		fmt.Printf("Move %s placed at coordinates: (%d, %d, %d)\n", moveInput, x, y, z)
		totalMoves++
		
		// Check for win
		winner := checkWin(&board)
		if winner != '|' {
			printBoard()
			fmt.Printf("\n🎉 %s wins! 🎉\n", playerNames[currentPlayer])
			return
		}
		
		// Switch to next player
		currentPlayer = (currentPlayer + 1) % 2
	}
	
	// If we reach here, it's a draw
	printBoard()
	fmt.Println("\n🤝 It's a draw! The board is full. 🤝")
}
