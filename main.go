package main

import "fmt"

var length int = 4
var width int = 4
var height int = 4
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

func move(move string, player byte, board *[4][4][4]byte) bool {
	// Move should be A3, B1, ..., L12, etc.
	if len(move) < 2 {
		return false
	}
	// Get column
	col := move[0] - 'A'
	if col < 0 || col >= byte(length) {
		return false
	}
	// Get row
	row := 0
	for i := 1; i < len(move); i++ {
		if move[i] < '0' || move[i] > '9' {
			return false
		}
		row = row*10 + int(move[i]-'0')
	}
	row--
	if row < 0 || row >= width {
		return false
	}
	// Find the lowest empty spot in the column
	for h := 0; h < height; h++ {
		if board[col][row][h] == '|' {
			board[col][row][h] = player
			return true
		}
	}
	return false
}

func main() {
	initBoard()
	printBoard()
	move("A1", 'x', &board)
	move("B2", 'o', &board)
	move("A1", 'x', &board)
	move("A1", 'o', &board)
	printBoard()
}
