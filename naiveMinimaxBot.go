package main

// NaiveMinimaxBot represents a simple minimax AI player without optimizations
type NaiveMinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int // Base for exponential scoring (e.g., 2, 3, 4)
}

// NewNaiveMinimaxBot creates a new naive minimax bot with the given symbol, name, and search depth
func NewNaiveMinimaxBot(symbol byte, name string, depth int, base int) *NaiveMinimaxBot {
	return &NaiveMinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
	}
}

// MakeMove makes a move using naive minimax algorithm (implements BotInterface)
// Uses full board evaluation at each step - no delta evaluation optimization
func (bot *NaiveMinimaxBot) MakeMove(board *Board) (string, [3]int) {
	_, bestMoves := naiveMinimax(board, bot.Depth, bot.Symbol == 'x')
	if len(bestMoves) == 0 {
		return "", [3]int{-1, -1, -1} // No valid moves
	}
	bestMove := bestMoves[0] // Pick the first best move
	coords := board.Move(bestMove, bot.Symbol)
	return bestMove, coords
}

// getName returns the bot's name (implements BotInterface)
func (bot *NaiveMinimaxBot) getName() string {
	return bot.Name
}

// getSymbol returns the bot's symbol (implements BotInterface)
func (bot *NaiveMinimaxBot) getSymbol() byte {
	return bot.Symbol
}

// naiveMinimax function uses full board evaluation instead of delta evaluation
func naiveMinimax(board *Board, depth int, isMaximizing bool) (int, []string) {
	// Check for winning conditions first
	winner := board.CheckWin()
	if winner != '|' {
		if winner == 'x' {
			return MAX_INT / 2, []string{} // X wins
		} else {
			return MIN_INT / 2, []string{} // O wins
		}
	}

	if depth == 0 {
		// Use full evaluation instead of cached score
		return board.Evaluate(), []string{}
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
		// Create a deep copy for naive approach (no move/unmove optimization)
		testBoard := copyBoard(board)
		testBoard.Move(move, symbol)

		score, moves := naiveMinimax(testBoard, depth-1, !isMaximizing)

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
