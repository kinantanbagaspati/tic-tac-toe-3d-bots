package main

import (
	"context"
	"sync"
	"sync/atomic"
)

// ConcurrentAlphaBetaMinimaxBot represents a concurrent minimax AI player with alpha-beta pruning
type ConcurrentAlphaBetaMinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int
	Base   int // Base for exponential scoring (e.g., 2, 3, 4)
}

// NewConcurrentAlphaBetaMinimaxBot creates a new concurrent alpha-beta minimax bot
func NewConcurrentAlphaBetaMinimaxBot(symbol byte, name string, depth int, base int) *ConcurrentAlphaBetaMinimaxBot {
	return &ConcurrentAlphaBetaMinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
		Base:   base,
	}
}

// MakeMove makes a move using concurrent alpha-beta pruning minimax algorithm (implements BotInterface)
func (bot *ConcurrentAlphaBetaMinimaxBot) MakeMove(board *Board) (string, [3]int) {
	// Use extreme threshold for root call (no pruning constraint from parent)
	isMaximizing := bot.Symbol == 'x'
	threshold := MIN_INT // If we're maximizing, use MIN_INT (can never prune)
	if !isMaximizing {
		threshold = MAX_INT // If we're minimizing, use MAX_INT (can never prune)
	}

	_, bestMoves := concurrentAlphaBetaMinimax(board, bot.Depth, isMaximizing, threshold, nil)
	if len(bestMoves) == 0 {
		return "", [3]int{-1, -1, -1} // No valid moves
	}
	bestMove := bestMoves[0] // Pick the first best move
	coords := board.Move(bestMove, bot.Symbol)
	return bestMove, coords
}

// getName returns the bot's name (implements BotInterface)
func (bot *ConcurrentAlphaBetaMinimaxBot) getName() string {
	return bot.Name
}

// getSymbol returns the bot's symbol (implements BotInterface)
func (bot *ConcurrentAlphaBetaMinimaxBot) getSymbol() byte {
	return bot.Symbol
}

// SharedScore represents a thread-safe score that can be updated and read by goroutines
type SharedScore struct {
	score int64 // Use int64 for atomic operations
}

// Get returns the current score
func (s *SharedScore) Get() int {
	return int(atomic.LoadInt64(&s.score))
}

// Update atomically updates the score if the new score is better
func (s *SharedScore) Update(newScore int, isMaximizing bool) bool {
	for {
		current := atomic.LoadInt64(&s.score)
		if (isMaximizing && newScore <= int(current)) || (!isMaximizing && newScore >= int(current)) {
			return false // New score is not better
		}
		if atomic.CompareAndSwapInt64(&s.score, current, int64(newScore)) {
			return true // Successfully updated
		}
		// Retry if another goroutine updated the score
	}
}

// concurrentAlphaBetaMinimax performs concurrent minimax with alpha-beta pruning
// parentScore: shared score that can be read to check for pruning conditions
func concurrentAlphaBetaMinimax(board *Board, depth int, isMaximizing bool, threshold int, parentScore *SharedScore) (int, []string) {
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
	if len(validMoves) <= 2 || depth <= 2 {
		return alphaBetaMinimax(board, depth, isMaximizing, threshold)
	}

	// Set result to very low/high initial value
	symbol := byte('x')
	initialScore := MIN_INT
	if !isMaximizing {
		symbol = 'o'
		initialScore = MAX_INT
	}

	// Shared score for this level - child goroutines will monitor this
	currentScore := &SharedScore{}
	atomic.StoreInt64(&currentScore.score, int64(initialScore))

	// Context for cancellation when pruning occurs
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Channel to collect results from goroutines
	results := make(chan MoveResult, len(validMoves))
	var wg sync.WaitGroup

	// Launch goroutines for each move
	for _, move := range validMoves {
		wg.Add(1)
		go func(move string) {
			defer wg.Done()

			// Check if we should terminate early due to parent's score update
			select {
			case <-ctx.Done():
				return // Parent found a better path, terminate
			default:
			}

			// Create a deep copy of the board to test the move
			testBoard := copyBoard(board)
			testBoard.Move(move, symbol)

			// Check pruning condition based on parent's current score
			if parentScore != nil {
				parentCurrentScore := parentScore.Get()
				if isMaximizing && parentCurrentScore <= threshold {
					// Parent is minimizing and found a score <= threshold, so it will prune us
					return
				}
				if !isMaximizing && parentCurrentScore >= threshold {
					// Parent is maximizing and found a score >= threshold, so it will prune us
					return
				}
			}

			// Pass current score as threshold for child nodes
			childThreshold := currentScore.Get()
			score, _ := concurrentAlphaBetaMinimax(testBoard, depth-1, !isMaximizing, childThreshold, currentScore)

			// Check if we've been cancelled while computing
			select {
			case <-ctx.Done():
				return
			default:
			}

			results <- MoveResult{Move: move, Score: score}
		}(move)
	}

	// Goroutine to close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results as they come in and update current score
	bestScore := initialScore
	bestMoves := []string{}

	for result := range results {
		if isMaximizing && result.Score > bestScore {
			bestScore = result.Score
			bestMoves = []string{result.Move} // Just store the immediate best move
			currentScore.Update(bestScore, isMaximizing)

			// Check if we can prune (our score beats the threshold)
			if bestScore >= threshold {
				cancel() // Signal other goroutines to terminate
				break
			}
		} else if !isMaximizing && result.Score < bestScore {
			bestScore = result.Score
			bestMoves = []string{result.Move} // Just store the immediate best move
			currentScore.Update(bestScore, isMaximizing)

			// Check if we can prune (our score beats the threshold)
			if bestScore <= threshold {
				cancel() // Signal other goroutines to terminate
				break
			}
		}
	}

	// Drain remaining results to prevent goroutine leaks
	for range results {
		// Consume any remaining results
	}

	return bestScore, bestMoves
}
