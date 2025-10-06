package main

// MinimaxBot represents a minimax AI player
type MinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int // Base for exponential scoring (e.g., 2, 3, 4)
}

// NewMinimaxBot creates a new minimax bot with the given symbol, name, and search depth
func NewMinimaxBot(symbol byte, name string, depth int, base int) *MinimaxBot {
	return &MinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
	}
}

// MakeMove makes a move using minimax algorithm (implements BotInterface)
// Currently uses evaluation function only - returns best evaluated move
func (bot *MinimaxBot) MakeMove(board *Board) (string, [3]int) {
	_, bestMoves := minimax(board, bot.Depth, bot.Symbol == 'x')
	if len(bestMoves) == 0 {
		return "", [3]int{-1, -1, -1} // No valid moves
	}
	bestMove := bestMoves[0] // Pick the first best move
	coords := board.Move(bestMove, bot.Symbol)
	return bestMove, coords
}

// getName returns the bot's name (implements BotInterface)
func (bot *MinimaxBot) getName() string {
	return bot.Name
}

// getSymbol returns the bot's symbol (implements BotInterface)
func (bot *MinimaxBot) getSymbol() byte {
	return bot.Symbol
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

// Default minimax function, returns pair of (score, array of best moves)
func minimax(board *Board, depth int, isMaximizing bool) (int, []string) {
	if depth == 0 {
		return board.Score, []string{} // Use the board's current score instead of recalculating
	}

	// Set result to very low/high initial value
	var symbol byte = 'x'
	bestScore := MIN_INT
	if !isMaximizing {
		symbol = 'o'
		bestScore = MAX_INT
	}
	bestMoves := []string{}

	for _, move := range board.GetValidMoves() {
		// Create a deep copy of the board to test the move
		testBoard := copyBoard(board)
		testBoard.Move(move, symbol)
		score, moves := minimax(testBoard, depth-1, !isMaximizing)
		if isMaximizing && score > bestScore {
			bestScore = score
			bestMoves = append([]string{move}, moves...)
		} else if !isMaximizing && score < bestScore {
			bestScore = score
			bestMoves = append([]string{move}, moves...)
		}
	}

	return bestScore, bestMoves
}
