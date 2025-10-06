package main

// AlphaBetaMinimaxBot represents a minimax AI player with threshold-based pruning optimization
type AlphaBetaMinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int // Base for exponential scoring (e.g., 2, 3, 4)
}

// NewAlphaBetaMinimaxBot creates a new threshold-based pruning minimax bot with the given symbol, name, and search depth
func NewAlphaBetaMinimaxBot(symbol byte, name string, depth int, base int) *AlphaBetaMinimaxBot {
	return &AlphaBetaMinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
	}
}

// MakeMove makes a move using alpha-beta pruning minimax algorithm (implements BotInterface)
// Uses threshold-based pruning to eliminate unnecessary branches from the search tree
func (bot *AlphaBetaMinimaxBot) MakeMove(board *Board) (string, [3]int) {
	// Use extreme threshold for root call (no pruning constraint from parent)
	isMaximizing := bot.Symbol == 'x'
	threshold := MIN_INT // If we're maximizing, use MIN_INT (can never prune)
	if !isMaximizing {
		threshold = MAX_INT // If we're minimizing, use MAX_INT (can never prune)
	}
	_, bestMoves := alphaBetaMinimax(board, bot.Depth, isMaximizing, threshold)
	if len(bestMoves) == 0 {
		return "", [3]int{-1, -1, -1} // No valid moves
	}
	bestMove := bestMoves[0] // Pick the first best move
	coords := board.Move(bestMove, bot.Symbol)
	return bestMove, coords
}

// getName returns the bot's name (implements BotInterface)
func (bot *AlphaBetaMinimaxBot) getName() string {
	return bot.Name
}

// getSymbol returns the bot's symbol (implements BotInterface)
func (bot *AlphaBetaMinimaxBot) getSymbol() byte {
	return bot.Symbol
}

// alphaBetaMinimax performs minimax with threshold-based pruning optimization
// This approach simplifies traditional alpha-beta pruning by using:
// - threshold: the current best score we're trying to beat (MAX_INT/MIN_INT if no constraint)
// When a score exceeds the threshold, we can prune the remaining search branches
func alphaBetaMinimax(board *Board, depth int, isMaximizing bool, threshold int) (int, []string) {
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
		return board.Score, []string{} // Use the board's current score
	}

	// Set result to very low/high initial value
	var symbol byte = 'x'
	currentScore := MIN_INT
	if !isMaximizing {
		symbol = 'o'
		currentScore = MAX_INT
	}
	bestMoves := []string{}

	for _, move := range board.GetValidMoves() {
		board.Move(move, symbol)

		// Pass our current best score as threshold for pruning
		score, moves := alphaBetaMinimax(board, depth-1, !isMaximizing, currentScore)
		board.UnMove(move)

		if isMaximizing {
			if score > currentScore {
				currentScore = score
				bestMoves = append([]string{move}, moves...)
			}
			// Threshold-based pruning: if our score beats the threshold, parent won't choose this path
			if currentScore >= threshold {
				break // Parent is minimizing and won't select this branch
			}
		} else {
			if score < currentScore {
				currentScore = score
				bestMoves = append([]string{move}, moves...)
			}
			// Threshold-based pruning: if our score is worse than threshold, parent won't choose this path
			if currentScore <= threshold {
				break // Parent is maximizing and won't select this branch
			}
		}
	}

	return currentScore, bestMoves
}
