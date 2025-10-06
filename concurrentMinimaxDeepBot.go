package main

import (
	"sync"
)

// ConcurrentMinimaxDeepBot represents a fully concurrent minimax AI player using goroutines at all levels
type ConcurrentMinimaxDeepBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int // Base for exponential scoring (e.g., 2, 3, 4)
}

// NewConcurrentMinimaxDeepBot creates a new deep concurrent minimax bot with the given symbol, name, and search depth
func NewConcurrentMinimaxDeepBot(symbol byte, name string, depth int, base int) *ConcurrentMinimaxDeepBot {
	return &ConcurrentMinimaxDeepBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
	}
}

// MakeMove makes a move using deep concurrent minimax algorithm (implements BotInterface)
// Uses concurrency at every level of the minimax tree
func (bot *ConcurrentMinimaxDeepBot) MakeMove(board *Board) (string, [3]int) {
	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return "", [3]int{-1, -1, -1} // No valid moves
	}

	// Use deep concurrent minimax to find the best move
	_, bestMoves := concurrentMinimaxDeep(board, bot.Depth, bot.Symbol == 'x')
	if len(bestMoves) == 0 {
		return "", [3]int{-1, -1, -1} // No valid moves
	}

	bestMove := bestMoves[0] // Pick the first best move
	coords := board.Move(bestMove, bot.Symbol)
	return bestMove, coords
}

// getName returns the bot's name (implements BotInterface)
func (bot *ConcurrentMinimaxDeepBot) getName() string {
	return bot.Name
}

// getSymbol returns the bot's symbol (implements BotInterface)
func (bot *ConcurrentMinimaxDeepBot) getSymbol() byte {
	return bot.Symbol
}

// concurrentMinimaxDeep performs fully concurrent minimax at every level
// This version uses goroutines at every level of the recursion for maximum parallelization
func concurrentMinimaxDeep(board *Board, depth int, isMaximizing bool) (int, []string) {
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

	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return board.Score, []string{} // Use the board's current score
	}

	// For small number of moves or shallow depth, use sequential to avoid overhead
	if len(validMoves) <= 2 || depth <= 1 {
		return minimax(board, depth, isMaximizing)
	}

	// Set result to very low/high initial value
	symbol := byte('x')
	if !isMaximizing {
		symbol = 'o'
	}

	// Channel to collect results from goroutines
	type DepthResult struct {
		Move  string
		Score int
		Moves []string
	}

	results := make(chan DepthResult, len(validMoves))
	var wg sync.WaitGroup

	for _, move := range validMoves {
		wg.Add(1)
		go func(move string) {
			defer wg.Done()

			// Create a deep copy of the board to test the move
			testBoard := copyBoard(board)
			testBoard.Move(move, symbol)

			// Recursively evaluate this branch with deep concurrency
			score, moves := concurrentMinimaxDeep(testBoard, depth-1, !isMaximizing)

			results <- DepthResult{Move: move, Score: score, Moves: moves}
		}(move)
	}

	// Close results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Find the best result from all branches
	bestScore := MIN_INT
	if !isMaximizing {
		bestScore = MAX_INT
	}
	bestMoves := []string{}

	for result := range results {
		if isMaximizing && result.Score > bestScore {
			bestScore = result.Score
			bestMoves = append([]string{result.Move}, result.Moves...)
		} else if !isMaximizing && result.Score < bestScore {
			bestScore = result.Score
			bestMoves = append([]string{result.Move}, result.Moves...)
		}
	}

	return bestScore, bestMoves
}
