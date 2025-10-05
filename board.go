package main

import "fmt"

// Board represents a 3D Tic-Tac-Toe board
type Board struct {
	Length         int
	Width          int
	Height         int
	WinLength      int
	Grid           [][][]byte
	CurrentHeights [][]int // Tracks the current height of each column [length][width]
	LastMove       [3]int  // Stores the last move coordinates [x, y, z], or [-1, -1, -1] if no moves yet
}

// NewBoard creates a new board with specified dimensions
// If no arguments provided, uses default dimensions (4x4x4, win=4)
// Usage: 
//   NewBoard() - creates 4x4x4 board with win=4
//   NewBoard(3) - creates 3x3x3 board with win=3
//   NewBoard(3, 3, 3, 3) - creates 3x3x3 board with win=3
func NewBoard(dimensions ...int) *Board {
	// Default values
	length, width, height, winLength := 4, 4, 4, 4
	
	// Override with provided values
	if len(dimensions) >= 1 {
		length = dimensions[0]
		width = dimensions[0]   // If only one dimension provided, use it for all
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
}

// Print displays the board in a 2D projection
func (b *Board) Print() {
	toPrint := make([][]byte, b.Length+b.Width+b.Height-2)
	for i := range toPrint {
		toPrint[i] = make([]byte, b.Length*b.Width)
		for j := range toPrint[i] {
			toPrint[i][j] = ' '
		}
	}

	for i := 0; i < b.Length; i++ {
		for j := 0; j < b.Width; j++ {
			for k := 0; k < b.Height; k++ {
				toPrint[i+b.Width-j+b.Height-k-2][i*b.Width+j] = b.Grid[i][j][k]
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
	if len(moveStr) < 2 {
		return [3]int{-1, -1, -1}
	}
	
	// Get column
	col := moveStr[0] - 'A'
	if col < 0 || col >= byte(b.Length) {
		return [3]int{-1, -1, -1}
	}
	
	// Get row
	row := 0
	for i := 1; i < len(moveStr); i++ {
		if moveStr[i] < '0' || moveStr[i] > '9' {
			return [3]int{-1, -1, -1}
		}
		row = row*10 + int(moveStr[i]-'0')
	}
	row--
	if row < 0 || row >= b.Width {
		return [3]int{-1, -1, -1}
	}
	
	// Try placing the block
	currentHeight := b.CurrentHeights[col][row]
	if currentHeight >= b.Height {
		return [3]int{-1, -1, -1}
	}
	b.Grid[col][row][currentHeight] = player
	b.CurrentHeights[col][row]++
	b.LastMove = [3]int{int(col), row, currentHeight}
	
	return b.LastMove
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

// CheckWin checks for a winning condition
// Returns 'x' if player X wins, 'o' if player O wins, or '|' if no winner
func (b *Board) CheckWin() byte {
	directions := [][3]int{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, // 1D
		{1, 1, 0}, {1, -1, 0}, {1, 0, 1}, {1, 0, -1}, {0, 1, 1}, {0, 1, -1}, // 2D diagonals
		{1, 1, 1}, {1, -1, -1}, {1, 1, -1}, {1, -1, 1}, // 3D diagonals
	}
	
	xPattern := make([]byte, b.WinLength)
	oPattern := make([]byte, b.WinLength)
	for p := 0; p < b.WinLength; p++ {
		xPattern[p] = 'x'
		oPattern[p] = 'o'
	}
	
	for i := 0; i < b.Length; i++ {
		for j := 0; j < b.Width; j++ {
			for k := 0; k < b.Height; k++ {
				// Check all directions from each cell
				for _, dir := range directions {
					if !b.IsValidCoordinate(i+(b.WinLength-1)*dir[0], j+(b.WinLength-1)*dir[1], k+(b.WinLength-1)*dir[2]) {
						continue
					}
					line := b.GetLine([3]int{i, j, k}, dir)
					
					if string(line) == string(xPattern) {
						for n := 0; n < b.WinLength; n++ {
							b.Grid[i+n*dir[0]][j+n*dir[1]][k+n*dir[2]] = 'X'
						}
						return 'x'
					}
					if string(line) == string(oPattern) {
						for n := 0; n < b.WinLength; n++ {
							b.Grid[i+n*dir[0]][j+n*dir[1]][k+n*dir[2]] = 'O'
						}
						return 'o'
					}
				}
			}
		}
	}
	return '|'
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
