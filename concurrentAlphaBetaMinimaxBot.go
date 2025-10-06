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

// MultiDepthStreamResult represents a streaming result from multiple depth evaluation
type MultiDepthStreamResult struct {
	Moves []string // sequence of moves
	Score int
	Depth int
	Final bool // true if this is the final result
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

// SequenceStreamResult represents a streaming result with move sequences
type SequenceStreamResult struct {
	Moves []string
	Score int
	Final bool
}

// concurrentAlphaBetaMinimaxStreamWithSequence performs streaming concurrent minimax that tracks move sequences
func concurrentAlphaBetaMinimaxStreamWithSequence(board *Board, depth int, isMaximizing bool, parentCtx context.Context) <-chan SequenceStreamResult {
	resultCh := make(chan SequenceStreamResult, 10)

	go func() {
		defer close(resultCh)

		// Check for winning conditions first
		winner := board.CheckWin()
		if winner != '|' {
			if winner == 'x' {
				resultCh <- SequenceStreamResult{Moves: []string{}, Score: MAX_INT / 2, Final: true}
			} else {
				resultCh <- SequenceStreamResult{Moves: []string{}, Score: MIN_INT / 2, Final: true}
			}
			return
		}

		if depth == 0 {
			resultCh <- SequenceStreamResult{Moves: []string{}, Score: board.Score, Final: true}
			return
		}

		validMoves := board.GetValidMoves()
		if len(validMoves) == 0 {
			resultCh <- SequenceStreamResult{Moves: []string{}, Score: board.Score, Final: true}
			return
		}

		// For small cases, use sequential
		if len(validMoves) <= 2 || depth <= 2 {
			threshold := MIN_INT
			if !isMaximizing {
				threshold = MAX_INT
			}
			score, moves := alphaBetaMinimax(board, depth, isMaximizing, threshold)
			resultCh <- SequenceStreamResult{Moves: moves, Score: score, Final: true}
			return
		}

		// Streaming concurrent evaluation with sequence tracking
		symbol := byte('x')
		bestScore := MIN_INT
		if !isMaximizing {
			symbol = 'o'
			bestScore = MAX_INT
		}

		var bestMoves []string

		// Context for cancellation
		if parentCtx == nil {
			parentCtx = context.Background()
		}
		ctx, cancel := context.WithCancel(parentCtx)
		defer cancel()

		// Channel to collect child results
		childResults := make(chan SequenceStreamResult, len(validMoves)*2)
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
				childCh := concurrentAlphaBetaMinimaxStreamWithSequence(testBoard, depth-1, !isMaximizing, ctx)

				// Forward all results from child, prepending current move
				for childResult := range childCh {
					select {
					case <-ctx.Done():
						return // Pruned by parent
					case childResults <- SequenceStreamResult{
						Moves: append([]string{move}, childResult.Moves...),
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
				bestMoves = result.Moves
				improved = true
			} else if !isMaximizing && result.Score < bestScore {
				bestScore = result.Score
				bestMoves = result.Moves
				improved = true
			}

			// Stream the improvement to parent
			if improved {
				select {
				case <-parentCtx.Done():
					return // Parent cancelled us
				case resultCh <- SequenceStreamResult{Moves: bestMoves, Score: bestScore, Final: false}:
				}

				// Check if we can prune remaining children
				if (isMaximizing && bestScore >= MAX_INT/3) || (!isMaximizing && bestScore <= MIN_INT/3) {
					cancel() // Signal children to stop
					break
				}
			}

			// If this was a final result for this move, mark it as complete
			if result.Final && len(result.Moves) > 0 {
				delete(activeMoves, result.Moves[0])

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
		case resultCh <- SequenceStreamResult{Moves: bestMoves, Score: bestScore, Final: true}:
		}
	}()

	return resultCh
}

// multiDepthAlphaBetaStream performs concurrent alpha-beta with multiple depths
// Returns a channel that streams the best moves found by different depth bots
func multiDepthAlphaBetaStream(board *Board, isMaximizing bool, depths []int) <-chan MultiDepthStreamResult {
	resultCh := make(chan MultiDepthStreamResult, 20) // Buffered for streaming

	go func() {
		defer close(resultCh)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Channel to collect results from all depth bots
		depthResults := make(chan MultiDepthStreamResult, len(depths)*5)
		var wg sync.WaitGroup

		// Launch a bot for each depth
		for _, depth := range depths {
			wg.Add(1)
			go func(depth int) {
				defer wg.Done()

				// Get streaming results from this depth
				streamCh := concurrentAlphaBetaMinimaxStreamWithSequence(board, depth, isMaximizing, ctx)

				// Forward results with depth information
				for result := range streamCh {
					select {
					case <-ctx.Done():
						return
					case depthResults <- MultiDepthStreamResult{
						Moves: result.Moves,
						Score: result.Score,
						Depth: depth,
						Final: result.Final,
					}:
					}
				}
			}(depth)
		}

		// Close results channel when all workers are done
		go func() {
			wg.Wait()
			close(depthResults)
		}()

		// Track best results and active depths
		bestScore := MIN_INT
		if !isMaximizing {
			bestScore = MAX_INT
		}
		var bestMoves []string
		bestDepth := 0
		activeDepths := make(map[int]bool)
		for _, depth := range depths {
			activeDepths[depth] = true
		}

		// Process results as they stream in
		for result := range depthResults {
			// Check if this result improves our best result
			// Priority: Deeper depth wins if scores are equal or better
			improved := false

			if isMaximizing {
				// For maximizing: better score OR (same/better score + deeper depth)
				if result.Depth > bestDepth ||
					(result.Score > bestScore && result.Depth == bestDepth) {
					bestScore = result.Score
					bestMoves = result.Moves
					bestDepth = result.Depth
					improved = true
				}
			} else {
				// For minimizing: better score OR (same/better score + deeper depth)
				if result.Depth > bestDepth ||
					(result.Score < bestScore && result.Depth == bestDepth) {
					bestScore = result.Score
					bestMoves = result.Moves
					bestDepth = result.Depth
					improved = true
				}
			}

			// Stream the improvement
			if improved {
				select {
				case <-ctx.Done():
					return
				case resultCh <- MultiDepthStreamResult{
					Moves: bestMoves,
					Score: bestScore,
					Depth: bestDepth,
					Final: false,
				}:
				}
			}

			// If this was a final result for this depth, mark it as complete
			if result.Final {
				delete(activeDepths, result.Depth)

				// If all depths are complete, send final result and exit
				if len(activeDepths) == 0 {
					select {
					case <-ctx.Done():
						return
					case resultCh <- MultiDepthStreamResult{
						Moves: bestMoves,
						Score: bestScore,
						Depth: bestDepth,
						Final: true,
					}:
					}
					return
				}
			}
		}
	}()

	return resultCh
}
