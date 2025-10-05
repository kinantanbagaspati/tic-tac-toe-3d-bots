package main

import (
	"math/rand"
	"time"
)

// Bot represents a simple AI player
type Bot struct {
	Symbol byte
	Name   string
}

// BotInterface defines the interface that all bots must implement
type BotInterface interface {
	MakeMove(board *Board) (string, [3]int)
}

// NewBot creates a new bot with the given symbol and name
func NewBot(symbol byte, name string) *Bot {
	return &Bot{
		Symbol: symbol,
		Name:   name,
	}
}

// MakeMove makes a random valid move on the board (implements BotInterface)
func (bot *Bot) MakeMove(board *Board) (string, [3]int) {
	return bot.MakeRandomMove(board)
}

// MakeRandomMove makes a random valid move on the board
func (bot *Bot) MakeRandomMove(board *Board) (string, [3]int) {
	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return "", [3]int{-1, -1, -1}
	}
	
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Pick a random valid move
	randomIndex := rand.Intn(len(validMoves))
	chosenMove := validMoves[randomIndex]
	
	// Make the move
	coords := board.Move(chosenMove, bot.Symbol)
	return chosenMove, coords
}
