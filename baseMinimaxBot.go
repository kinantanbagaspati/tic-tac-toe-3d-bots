package main

// MinimaxBot represents a minimax AI player
type MinimaxBot struct {
	Symbol byte
	Name   string
	Depth  int // Search depth for minimax algorithm
}

// NewMinimaxBot creates a new minimax bot with the given symbol, name, and search depth
func NewMinimaxBot(symbol byte, name string, depth int) *MinimaxBot {
	return &MinimaxBot{
		Symbol: symbol,
		Name:   name,
		Depth:  depth,
	}
}



// MakeMove makes a move using minimax algorithm (implements BotInterface)
// Currently empty implementation - returns invalid move
func (bot *MinimaxBot) MakeMove(board *Board) (string, [3]int) {
	// TODO: Implement minimax algorithm
	// For now, return an invalid move to satisfy the interface
	return "", [3]int{-1, -1, -1}
}

// TODO: Add minimax algorithm implementation
// - evaluateBoard(board *Board) int
// - minimax(board *Board, depth int, isMaximizing bool, alpha int, beta int) int
// - getBestMove(board *Board) (string, [3]int)
