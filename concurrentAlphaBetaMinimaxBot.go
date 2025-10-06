package main

import (
	"context"
	"sync"
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

// MakeMove makes a move using streaming concurrent alpha-beta pruning minimax algorithm (implements BotInterface)
func (bot *ConcurrentAlphaBetaMinimaxBot) MakeMove(board *Board) (string, [3]int) {
	// Use streaming concurrent minimax
	resultCh := concurrentAlphaBetaMinimaxStream(board, bot.Depth, bot.Symbol == 'x', context.Background())

	var bestMove string

	// Listen to the stream until we get the final result
	for result := range resultCh {
		if result.Final {
			bestMove = result.Move
			break
		}
		// Keep updating with better moves as they're found
		bestMove = result.Move
	}

	if bestMove == "" {
		return "", [3]int{-1, -1, -1} // No valid moves
	}

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

// StreamResult represents a streaming result from minimax evaluation
type StreamResult struct {
	Move  string
	Score int
	Final bool // true if this is the final result for this branch
}

// concurrentAlphaBetaMinimaxStream performs streaming concurrent minimax with alpha-beta pruning
// Returns a channel that continuously emits better moves as they're discovered
func concurrentAlphaBetaMinimaxStream(board *Board, depth int, isMaximizing bool, parentCtx context.Context) <-chan StreamResult {
	resultCh := make(chan StreamResult, 10) // Buffered for streaming

	go func() {
		defer close(resultCh)

		// Check for winning conditions first
		winner := board.CheckWin()
		if winner != '|' {
			if winner == 'x' {
				resultCh <- StreamResult{Move: "", Score: MAX_INT / 2, Final: true}
			} else {
				resultCh <- StreamResult{Move: "", Score: MIN_INT / 2, Final: true}
			}
			return
		}

		if depth == 0 {
			resultCh <- StreamResult{Move: "", Score: board.Score, Final: true}
			return
		}

		validMoves := board.GetValidMoves()
		if len(validMoves) == 0 {
			resultCh <- StreamResult{Move: "", Score: board.Score, Final: true}
			return
		}

		// For small cases, use sequential to avoid overhead
		if len(validMoves) <= 2 || depth <= 2 {
			threshold := MIN_INT
			if !isMaximizing {
				threshold = MAX_INT
			}
			score, moves := alphaBetaMinimax(board, depth, isMaximizing, threshold)
			move := ""
			if len(moves) > 0 {
				move = moves[0]
			}
			resultCh <- StreamResult{Move: move, Score: score, Final: true}
			return
		}

		// Streaming concurrent evaluation
		symbol := byte('x')
		bestScore := MIN_INT
		if !isMaximizing {
			symbol = 'o'
			bestScore = MAX_INT
		}

		var bestMove string

		// Context for cancellation
		if parentCtx == nil {
			parentCtx = context.Background()
		}
		ctx, cancel := context.WithCancel(parentCtx)
		defer cancel()

		// Channel to collect child results
		childResults := make(chan StreamResult, len(validMoves)*2) // Buffer for multiple results per child
		var wg sync.WaitGroup

		// Launch goroutines for each move
		for _, move := range validMoves {
			wg.Add(1)
			go func(move string) {
				defer wg.Done()

				// Create a deep copy for this move
				testBoard := copyBoard(board)
				testBoard.Move(move, symbol)

				// Start streaming evaluation for this child
				childCh := concurrentAlphaBetaMinimaxStream(testBoard, depth-1, !isMaximizing, ctx)

				// Forward all results from child, tagging with the move
				for childResult := range childCh {
					select {
					case <-ctx.Done():
						return // Pruned by parent
					case childResults <- StreamResult{
						Move:  move,
						Score: childResult.Score,
						Final: childResult.Final,
					}:
					}

					// Stop if child sent final result
					if childResult.Final {
						break
					}
				}
			}(move)
		}

		// Close results channel when all workers are done
		go func() {
			wg.Wait()
			close(childResults)
		}()

		// Process streaming results from all children
		activeMoves := make(map[string]bool)
		for _, move := range validMoves {
			activeMoves[move] = true
		}

		for result := range childResults {
			// Check if this result improves our best score
			improved := false
			if isMaximizing && result.Score > bestScore {
				bestScore = result.Score
				bestMove = result.Move
				improved = true
			} else if !isMaximizing && result.Score < bestScore {
				bestScore = result.Score
				bestMove = result.Move
				improved = true
			}

			// Stream the improvement to parent
			if improved {
				select {
				case <-parentCtx.Done():
					return // Parent cancelled us
				case resultCh <- StreamResult{Move: bestMove, Score: bestScore, Final: false}:
				}

				// Check if we can prune remaining children (using reasonable thresholds)
				if (isMaximizing && bestScore >= MAX_INT/3) || (!isMaximizing && bestScore <= MIN_INT/3) {
					cancel() // Signal children to stop
					break
				}
			}

			// If this was a final result for this move, mark it as complete
			if result.Final {
				delete(activeMoves, result.Move)

				// If all moves are complete, we're done
				if len(activeMoves) == 0 {
					break
				}
			}
		}

		// Send final result
		select {
		case <-parentCtx.Done():
			return
		case resultCh <- StreamResult{Move: bestMove, Score: bestScore, Final: true}:
		}
	}()

	return resultCh
}
