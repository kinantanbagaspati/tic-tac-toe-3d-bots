package main

import (
	"fmt"
	"math"
)

// Board represents a 3D Tic-Tac-Toe board
type Board struct {
	Length         int
	Width          int
	Height         int
	WinLength      int
	Grid           [][][]byte
	CurrentHeights [][]int // Tracks the current height of each column [length][width]
	LastMove       [3]int  // Stores the last move coordinates [x, y, z], or [-1, -1, -1] if no moves yet
	Score          int     // Current board evaluation score (+ favors 'x', - favors 'o')
	Base           int     // Base for exponential scoring (e.g., 3, 10)
	PlayerWin      byte    // Stores who wins: 'x', 'o', or '|' for no winner
}

// NewBoard creates a new board with specified dimensions
// If no arguments provided, uses default dimensions (4x4x4, win=4)
// Usage:
//
//	NewBoard() - creates 4x4x4 board with win=4
//	NewBoard(3) - creates 3x3x3 board with win=3
//	NewBoard(3, 3, 3, 3) - creates 3x3x3 board with win=3
func NewBoard(dimensions ...int) *Board {
	// Default values
	length, width, height, winLength, base := 4, 4, 4, 4, 10

	// Override with provided values
	if len(dimensions) >= 1 {
		length = dimensions[0]
		width = dimensions[0] // If only one dimension provided, use it for all
		height = dimensions[0]
		winLength = dimensions[0]
	}
	if len(dimensions) >= 4 {
		length = dimensions[0]
		width = dimensions[1]
		height = dimensions[2]
		winLength = dimensions[3]
	}

	b := &Board{
		Length:    length,
		Width:     width,
		Height:    height,
		WinLength: winLength,
		Score:     0, // Start with neutral score
		Base:      base,
	}
	b.Init()
	return b
}

// Init initializes the board with empty markers
func (b *Board) Init() {
	// Initialize the 3D grid
	b.Grid = make([][][]byte, b.Length)
	for i := 0; i < b.Length; i++ {
		b.Grid[i] = make([][]byte, b.Width)
		for j := 0; j < b.Width; j++ {
			b.Grid[i][j] = make([]byte, b.Height)
			for k := 0; k < b.Height; k++ {
				b.Grid[i][j][k] = '|'
			}
		}
	}

	// Initialize the height tracking array
	b.CurrentHeights = make([][]int, b.Length)
	for i := 0; i < b.Length; i++ {
		b.CurrentHeights[i] = make([]int, b.Width)
		// Heights start at 0 (all columns are empty)
	}

	// Initialize last move to indicate no moves yet
	b.LastMove = [3]int{-1, -1, -1}

	// Initialize player win to no winner
	b.PlayerWin = '|'
}

// copyBoard creates a deep copy of the board for testing moves
func copyBoard(original *Board) *Board {
	// Create new board with same dimensions and evaluation base
	newBoard := NewBoard(original.Length, original.Width, original.Height, original.WinLength, original.Base)

	// Copy the grid state
	for i := 0; i < original.Length; i++ {
		for j := 0; j < original.Width; j++ {
			for k := 0; k < original.Height; k++ {
				newBoard.Grid[i][j][k] = original.Grid[i][j][k]
			}
		}
	}

	// Copy the height tracking
	for i := 0; i < original.Length; i++ {
		for j := 0; j < original.Width; j++ {
			newBoard.CurrentHeights[i][j] = original.CurrentHeights[i][j]
		}
	}

	// Copy last move, score, and player win
	newBoard.LastMove = original.LastMove
	newBoard.Score = original.Score
	newBoard.PlayerWin = original.PlayerWin

	return newBoard
}

// parseMove extracts column and row from move string (e.g., "A1" -> col=0, row=0)
// Returns (-1, -1) if the move string is invalid
func parseMove(moveStr string) (int, int) {
	if len(moveStr) < 2 {
		return -1, -1
	}

	// Get column
	col := int(moveStr[0]) - int('A')

	// Get row
	row := 0
	for i := 1; i < len(moveStr); i++ {
		if moveStr[i] < '0' || moveStr[i] > '9' {
			return -1, -1
		}
		row = row*10 + int(moveStr[i]-'0')
	}
	row-- // Convert from 1-based to 0-based indexing

	return col, row
}

// Print displays the board in a 2D projection
// Shows winning lines and check threats with capital letters and '#' for critical cells
func (b *Board) Print() {
	toPrint := make([][]byte, b.Length+b.Width+b.Height-2)
	for i := range toPrint {
		toPrint[i] = make([]byte, b.Length*b.Width)
		for j := range toPrint[i] {
			toPrint[i][j] = ' '
		}
	}

	// First, fill in the normal board state
	for i := 0; i < b.Length; i++ {
		for j := 0; j < b.Width; j++ {
			for k := 0; k < b.Height; k++ {
				toPrint[i+b.Width-j+b.Height-k-2][i*b.Width+j] = b.Grid[i][j][k]
			}
		}
	}

	directions := [][3]int{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, // 1D
		{1, 1, 0}, {1, -1, 0}, {1, 0, 1}, {1, 0, -1}, {0, 1, 1}, {0, 1, -1}, // 2D diagonals
		{1, 1, 1}, {1, -1, -1}, {1, 1, -1}, {1, -1, 1}, // 3D diagonals
	}

	// Check all lines for winning conditions and check threats
	for i := 0; i < b.Length; i++ {
		for j := 0; j < b.Width; j++ {
			for k := 0; k < b.Height; k++ {
				for _, dir := range directions {
					// Check if this line segment is valid
					endX := i + (b.WinLength-1)*dir[0]
					endY := j + (b.WinLength-1)*dir[1]
					endZ := k + (b.WinLength-1)*dir[2]

					if !b.IsValidCoordinate(endX, endY, endZ) {
						continue
					}

					line := b.GetLine([3]int{i, j, k}, dir)
					xCount := countBytes(line, 'x')
					oCount := countBytes(line, 'o')
					emptyCount := countBytes(line, '|')

					// Case 1: Winning line (all pieces of one player)
					if (xCount == b.WinLength) || (oCount == b.WinLength) {
						// Highlight all pieces in winning line as capitals
						for pos := 0; pos < b.WinLength; pos++ {
							x := i + pos*dir[0]
							y := j + pos*dir[1]
							z := k + pos*dir[2]

							printY := x + b.Width - y + b.Height - z - 2
							printX := x*b.Width + y

							if printY >= 0 && printY < len(toPrint) && printX >= 0 && printX < len(toPrint[printY]) {
								currentPiece := toPrint[printY][printX]
								if currentPiece == 'x' {
									toPrint[printY][printX] = 'X'
								} else if currentPiece == 'o' {
									toPrint[printY][printX] = 'O'
								}
							}
						}
					}

					// Case 2: Check threat (winLength-1 pieces + 1 empty that can be played)
					if emptyCount == 1 && (oCount == 0 || xCount == 0) {
						var criticalCell [3]int

						// Find the empty cell
						for pos := 0; pos < b.WinLength; pos++ {
							x := i + pos*dir[0]
							y := j + pos*dir[1]
							z := k + pos*dir[2]

							if b.Grid[x][y][z] == '|' {
								criticalCell = [3]int{x, y, z}
								break
							}
						}

						// Check if the critical cell can actually be played (correct height)
						canBePlayed := (criticalCell[2] == b.CurrentHeights[criticalCell[0]][criticalCell[1]])

						if canBePlayed {
							// Highlight threat line pieces in capital letters
							for pos := 0; pos < b.WinLength; pos++ {
								x := i + pos*dir[0]
								y := j + pos*dir[1]
								z := k + pos*dir[2]

								printY := x + b.Width - y + b.Height - z - 2
								printX := x*b.Width + y

								if printY >= 0 && printY < len(toPrint) && printX >= 0 && printX < len(toPrint[printY]) {
									currentPiece := toPrint[printY][printX]
									if currentPiece == 'x' {
										toPrint[printY][printX] = 'X'
									} else if currentPiece == 'o' {
										toPrint[printY][printX] = 'O'
									} else if currentPiece == '|' {
										toPrint[printY][printX] = '#'
									}
								}
							}
						}
					}
				}
			}
		}
	}

	for i := range toPrint {
		fmt.Println(string(toPrint[i]))
	}
}

// Move places a player's piece at the specified position
// Returns the coordinates where the piece was placed as [3]int, or [-1, -1, -1] if invalid
func (b *Board) Move(moveStr string, player byte) [3]int {
	// Parse the move string
	col, row := parseMove(moveStr)
	if col < 0 || col >= b.Length || row < 0 || row >= b.Width {
		return [3]int{-1, -1, -1}
	}

	// Try placing the block
	currentHeight := b.CurrentHeights[col][row]
	if currentHeight >= b.Height {
		return [3]int{-1, -1, -1}
	}

	// Place the piece first
	b.Grid[col][row][currentHeight] = player
	b.CurrentHeights[col][row]++
	b.LastMove = [3]int{col, row, currentHeight}

	// Calculate score delta after placing the piece and update win status
	delta := b.DeltaEvaluate(col, row, currentHeight, true)

	// Update the board's score with the delta
	b.Score += delta

	return b.LastMove
}

// UnMove reverses a move at the given position by removing the topmost piece
// and updating the score accordingly
func (b *Board) UnMove(moveStr string) [3]int {
	// Parse the move string
	col, row := parseMove(moveStr)
	if col < 0 || col >= b.Length || row < 0 || row >= b.Width {
		return [3]int{-1, -1, -1}
	}

	// Check if there's a piece to remove
	currentHeight := b.CurrentHeights[col][row]
	if currentHeight <= 0 {
		return [3]int{-1, -1, -1}
	}

	// Get the height of the topmost piece (0-based)
	topHeight := currentHeight - 1

	// Calculate the delta before removing the piece (don't update win status)
	delta := b.DeltaEvaluate(col, row, topHeight, false)

	// Remove the piece
	b.Grid[col][row][topHeight] = '|'
	b.CurrentHeights[col][row]--

	// Reverse the score delta and reset win status
	b.Score -= delta
	b.PlayerWin = '|'

	return [3]int{col, row, topHeight}
}

// IsValidCoordinate checks if the given coordinates are within board bounds
func (b *Board) IsValidCoordinate(x, y, z int) bool {
	return x >= 0 && x < b.Length && y >= 0 && y < b.Width && z >= 0 && z < b.Height
}

// GetLine returns a line of pieces starting from a position in a given direction
func (b *Board) GetLine(start [3]int, direction [3]int) []byte {
	line := make([]byte, b.WinLength)
	for i := 0; i < b.WinLength; i++ {
		x := start[0] + i*direction[0]
		y := start[1] + i*direction[1]
		z := start[2] + i*direction[2]
		if !b.IsValidCoordinate(x, y, z) {
			// Return a slice filled with invalid markers
			invalidLine := make([]byte, b.WinLength)
			for j := range invalidLine {
				invalidLine[j] = '|'
			}
			return invalidLine
		}
		line[i] = b.Grid[x][y][z]
	}
	return line
}

// CheckWin returns the current winner stored in PlayerWin field
// Returns 'x' if player X wins, 'o' if player O wins, or '|' if no winner
func (b *Board) CheckWin() byte {
	return b.PlayerWin
}

// GetValidMoves returns a slice of all valid move positions
func (b *Board) GetValidMoves() []string {
	var validMoves []string
	for i := 0; i < b.Length; i++ {
		for j := 0; j < b.Width; j++ {
			if b.CurrentHeights[i][j] < b.Height {
				move := fmt.Sprintf("%c%d", 'A'+byte(i), j+1)
				validMoves = append(validMoves, move)
			}
		}
	}
	return validMoves
}

// IsFull checks if the board is completely filled
func (b *Board) IsFull() bool {
	return len(b.GetValidMoves()) == 0
}

// Evaluate calculates the full board evaluation score
// + is good for 'x', - is good for 'o'
func (b *Board) Evaluate() int {
	directions := [][3]int{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, // 1D
		{1, 1, 0}, {1, -1, 0}, {1, 0, 1}, {1, 0, -1}, {0, 1, 1}, {0, 1, -1}, // 2D diagonals
		{1, 1, 1}, {1, -1, -1}, {1, 1, -1}, {1, -1, 1}, // 3D diagonals
	}
	score := 0

	for i := 0; i < b.Length; i++ {
		for j := 0; j < b.Width; j++ {
			for k := 0; k < b.Height; k++ {
				// Check all directions from each cell
				for _, dir := range directions {
					if !b.IsValidCoordinate(i+(b.WinLength-1)*dir[0], j+(b.WinLength-1)*dir[1], k+(b.WinLength-1)*dir[2]) {
						continue
					}
					line := b.GetLine([3]int{i, j, k}, dir)
					xCount := countBytes(line, 'x')
					oCount := countBytes(line, 'o')

					if xCount > 0 && oCount == 0 && xCount <= b.WinLength {
						score += int(math.Pow(float64(b.Base), float64(xCount)))
					} else if oCount > 0 && xCount == 0 && oCount <= b.WinLength {
						score -= int(math.Pow(float64(b.Base), float64(oCount)))
					}
				}
			}
		}
	}

	b.Score = score // Update the board's score
	return score
}

// DeltaEvaluate calculates the change in evaluation score for a piece at the given coordinates
// The piece must already be placed on the board. This is much more efficient than recalculating the entire board
// If updateWin is true, it will check for and update the PlayerWin field when a win is detected
func (b *Board) DeltaEvaluate(x, y, z int, updateWin bool) int {
	directions := [][3]int{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, // 1D
		{1, 1, 0}, {1, -1, 0}, {1, 0, 1}, {1, 0, -1}, {0, 1, 1}, {0, 1, -1}, // 2D diagonals
		{1, 1, 1}, {1, -1, -1}, {1, 1, -1}, {1, -1, 1}, // 3D diagonals
	}

	// Get the symbol of the piece at this position
	symbol := b.Grid[x][y][z]
	delta := 0

	// For each direction, check all lines that pass through this position
	for _, dir := range directions {
		// Check lines in both directions from this point
		for offset := -(b.WinLength - 1); offset <= 0; offset++ {
			startX := x + offset*dir[0]
			startY := y + offset*dir[1]
			startZ := z + offset*dir[2]

			endX := startX + (b.WinLength-1)*dir[0]
			endY := startY + (b.WinLength-1)*dir[1]
			endZ := startZ + (b.WinLength-1)*dir[2]

			// Check if this line segment is valid
			if !b.IsValidCoordinate(startX, startY, startZ) || !b.IsValidCoordinate(endX, endY, endZ) {
				continue
			}

			// Get the current line (with the piece already placed)
			lineAfter := b.GetLine([3]int{startX, startY, startZ}, dir)
			xCountAfter := countBytes(lineAfter, 'x')
			oCountAfter := countBytes(lineAfter, 'o')

			// Check for winning conditions and update PlayerWin if requested
			if updateWin && xCountAfter == b.WinLength && oCountAfter == 0 {
				b.PlayerWin = 'x'
			} else if updateWin && oCountAfter == b.WinLength && xCountAfter == 0 {
				b.PlayerWin = 'o'
			}

			// Calculate score contribution with the piece
			scoreAfter := 0
			if xCountAfter > 0 && oCountAfter == 0 && xCountAfter <= b.WinLength {
				scoreAfter += int(math.Pow(float64(b.Base), float64(xCountAfter)))
			} else if oCountAfter > 0 && xCountAfter == 0 && oCountAfter <= b.WinLength {
				scoreAfter -= int(math.Pow(float64(b.Base), float64(oCountAfter)))
			}

			// Calculate what the counts were before the move
			var xCountBefore, oCountBefore int
			if symbol == 'x' {
				xCountBefore = xCountAfter - 1
				oCountBefore = oCountAfter
			} else if symbol == 'o' {
				xCountBefore = xCountAfter
				oCountBefore = oCountAfter - 1
			} else {
				// Invalid symbol, skip this calculation
				continue
			}

			// Calculate score contribution before the move
			scoreBefore := 0
			if xCountBefore > 0 && oCountBefore == 0 && xCountBefore <= b.WinLength {
				scoreBefore += int(math.Pow(float64(b.Base), float64(xCountBefore)))
			} else if oCountBefore > 0 && xCountBefore == 0 && oCountBefore <= b.WinLength {
				scoreBefore -= int(math.Pow(float64(b.Base), float64(oCountBefore)))
			}

			// Add the delta for this line
			delta += scoreAfter - scoreBefore
		}
	}

	return delta
}
