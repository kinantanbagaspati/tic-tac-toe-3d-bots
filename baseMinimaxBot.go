package main

// MinimaxBot represents a minimax AI player
type MinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int   // Base for exponential scoring (e.g., 2, 3, 4)
	Powers []int // Precomputed powers: [base^0, base^1, base^2, ...]
}

// NewMinimaxBot creates a new minimax bot with the given symbol, name, and search depth
func NewMinimaxBot(symbol byte, name string, depth int, base int) *MinimaxBot {
	// Precompute powers up to base^winLength (we need up to winLength powers)
	// For safety, compute a few extra powers
	maxPower := 6 // Should cover most reasonable win lengths
	powers := make([]int, maxPower+1)
	powers[0] = 1 // base^0 = 1
	for i := 1; i <= maxPower; i++ {
		powers[i] = powers[i-1] * base // base^i = base^(i-1) * base
	}

	return &MinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
		Powers: powers,
	}
}

// MakeMove makes a move using minimax algorithm (implements BotInterface)
// Currently uses evaluation function only - returns best evaluated move
func (bot *MinimaxBot) MakeMove(board *Board) (string, [3]int) {
	// TODO: Implement full minimax algorithm
	// For now, use simple evaluation to pick the best immediate move

	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return "", [3]int{-1, -1, -1}
	}

	bestMove := validMoves[0]
	bestScore := -999999 // Very low initial score

	// Try each valid move and pick the one with best evaluation
	for _, move := range validMoves {
		// Create a deep copy of the board to test the move
		testBoard := copyBoard(board)
		coords := testBoard.Move(move, bot.Symbol)
		if coords[0] != -1 { // Valid move
			score := EvalExpo(testBoard, bot.Powers)
			// If bot is 'o', we want lower scores (negative is good for 'o')
			if bot.Symbol == 'o' {
				score = -score
			}
			if score > bestScore {
				bestScore = score
				bestMove = move
			}
		}
	}

	// Make the actual move
	return bestMove, board.Move(bestMove, bot.Symbol)
}

// copyBoard creates a deep copy of the board for testing moves
func copyBoard(original *Board) *Board {
	// Create new board with same dimensions
	newBoard := NewBoard(original.Length, original.Width, original.Height, original.WinLength)

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

	// Copy last move
	newBoard.LastMove = original.LastMove

	return newBoard
}

// countBytes counts how many times target appears in the byte slice
func countBytes(bytes []byte, target byte) int {
	count := 0
	for _, b := range bytes {
		if b == target {
			count++
		}
	}
	return count
}

func EvalExpo(board *Board, powers []int) int {
	// + is good for 'x', - is good for 'o'
	directions := [][3]int{
		{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, // 1D
		{1, 1, 0}, {1, -1, 0}, {1, 0, 1}, {1, 0, -1}, {0, 1, 1}, {0, 1, -1}, // 2D diagonals
		{1, 1, 1}, {1, -1, -1}, {1, 1, -1}, {1, -1, 1}, // 3D diagonals
	}
	score := 0

	for i := 0; i < board.Length; i++ {
		for j := 0; j < board.Width; j++ {
			for k := 0; k < board.Height; k++ {
				// Check all directions from each cell
				for _, dir := range directions {
					if !board.IsValidCoordinate(i+(board.WinLength-1)*dir[0], j+(board.WinLength-1)*dir[1], k+(board.WinLength-1)*dir[2]) {
						continue
					}
					line := board.GetLine([3]int{i, j, k}, dir)
					xCount := countBytes(line, 'x')
					oCount := countBytes(line, 'o')

					if xCount > 0 && oCount == 0 && xCount < len(powers) {
						score += powers[xCount]
					} else if oCount > 0 && xCount == 0 && oCount < len(powers) {
						score -= powers[oCount]
					}
				}
			}
		}
	}
	return score
}

// - evaluateBoard(board *Board) int
// - minimax(board *Board, depth int, isMaximizing bool, alpha int, beta int) int
// - getBestMove(board *Board) (string, [3]int)
