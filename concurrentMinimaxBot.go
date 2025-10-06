package main

import (
	"sync"
)

// ConcurrentMinimaxBot represents a concurrent minimax AI player using goroutines
type ConcurrentMinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int // Base for exponential scoring (e.g., 2, 3, 4)
}

// NewConcurrentMinimaxBot creates a new concurrent minimax bot with the given symbol, name, and search depth
func NewConcurrentMinimaxBot(symbol byte, name string, depth int, base int) *ConcurrentMinimaxBot {
	return &ConcurrentMinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
	}
}

// MoveResult represents the result of evaluating a move
type MoveResult struct {
	Move  string
	Score int
}

// MakeMove makes a move using concurrent minimax algorithm (implements BotInterface)
func (bot *ConcurrentMinimaxBot) MakeMove(board *Board) (string, [3]int) {
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
func (bot *ConcurrentMinimaxBot) getName() string {
	return bot.Name
}

// getSymbol returns the bot's symbol (implements BotInterface)
func (bot *ConcurrentMinimaxBot) getSymbol() byte {
	return bot.Symbol
}

// concurrentMinimax evaluates all possible moves concurrently and returns the best one
func concurrentMinimax(board *Board, depth int, isMaximizing bool, validMoves []string) string {
	if len(validMoves) == 0 {
		return ""
	}

	// If only one move available, return it immediately
	if len(validMoves) == 1 {
		return validMoves[0]
	}

	// Channel to collect results from goroutines
	results := make(chan MoveResult, len(validMoves))
	var wg sync.WaitGroup

	// Evaluate each possible move concurrently
	symbol := byte('x')
	if !isMaximizing {
		symbol = 'o'
	}

	for _, move := range validMoves {
		wg.Add(1)
		go func(move string) {
			defer wg.Done()

			// Create a deep copy of the board to test the move
			testBoard := copyBoard(board)
			testBoard.Move(move, symbol)

			// Evaluate this move using sequential minimax from this point
			score, _ := minimax(testBoard, depth-1, !isMaximizing)

			results <- MoveResult{Move: move, Score: score}
		}(move)
	}

	// Close results channel when all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Find the best move from all results
	bestScore := MIN_INT
	if !isMaximizing {
		bestScore = MAX_INT
	}
	bestMove := validMoves[0] // Default to first move

	for result := range results {
		if isMaximizing && result.Score > bestScore {
			bestScore = result.Score
			bestMove = result.Move
		} else if !isMaximizing && result.Score < bestScore {
			bestScore = result.Score
			bestMove = result.Move
		}
	}

	return bestMove
}

// concurrentMinimaxDeep performs fully concurrent minimax (alternative implementation)
// This version uses goroutines at every level of the recursion
func concurrentMinimaxDeep(board *Board, depth int, isMaximizing bool) (int, []string) {
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

			// Recursively evaluate this branch
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
