package main

import (
	"fmt"
	"math/rand"
	"time"
)

// Bot represents a simple AI player
type Bot struct {
	Symbol byte
	Name   string
}

// NewBot creates a new bot with the given symbol and name
func NewBot(symbol byte, name string) *Bot {
	return &Bot{
		Symbol: symbol,
		Name:   name,
	}
}

// MakeRandomMove makes a random valid move on the board
func (bot *Bot) MakeRandomMove(board *Board) (string, int, int, int) {
	validMoves := board.GetValidMoves()
	if len(validMoves) == 0 {
		return "", -1, -1, -1
	}
	
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())
	
	// Pick a random valid move
	randomIndex := rand.Intn(len(validMoves))
	chosenMove := validMoves[randomIndex]
	
	// Make the move
	x, y, z := board.Move(chosenMove, bot.Symbol)
	return chosenMove, x, y, z
}

// RunPvE starts a Player vs Environment (Bot) game
func RunPvE() {
	board := NewBoard(4, 4, 4, 4)
	bot := NewBot('o', "RandomBot")
	
	totalMoves := 0
	maxMoves := board.Length * board.Width * board.Height
	
	fmt.Println("🤖 Player vs Bot Mode")
	fmt.Println("Welcome to 3D Tic-Tac-Toe!")
	fmt.Printf("You are 'x', Bot is 'o'\n")
	fmt.Printf("Enter moves in format like A1, B2, etc. (A-%c, 1-%d)\n", 'A'+byte(board.Length-1), board.Width)
	fmt.Println()
	
	for totalMoves < maxMoves {
		board.Print()
		
		// Player's turn
		fmt.Printf("\nYour turn (playing 'x'): ")
		var moveInput string
		fmt.Scanln(&moveInput)
		
		x, y, z := board.Move(moveInput, 'x')
		if x == -1 && y == -1 && z == -1 {
			fmt.Println("Invalid move! Try again.")
			continue
		}
		
		fmt.Printf("Your move %s placed at coordinates: (%d, %d, %d)\n", moveInput, x, y, z)
		totalMoves++
		
		// Check for player win
		winner := board.CheckWin()
		if winner == 'x' {
			board.Print()
			fmt.Printf("\n🎉 You win! 🎉\n")
			return
		}
		
		// Check if board is full
		if board.IsFull() {
			break
		}
		
		// Bot's turn
		fmt.Printf("\n%s is thinking...\n", bot.Name)
		time.Sleep(1 * time.Second) // Add some delay for dramatic effect
		
		botMove, bx, by, bz := bot.MakeRandomMove(board)
		if bx == -1 && by == -1 && bz == -1 {
			break // No valid moves left
		}
		
		fmt.Printf("%s plays %s at coordinates: (%d, %d, %d)\n", bot.Name, botMove, bx, by, bz)
		totalMoves++
		
		// Check for bot win
		winner = board.CheckWin()
		if winner == 'o' {
			board.Print()
			fmt.Printf("\n🤖 %s wins! Better luck next time! 🤖\n", bot.Name)
			return
		}
		
		// Check if board is full
		if board.IsFull() {
			break
		}
	}
	
	// If we reach here, it's a draw
	board.Print()
	fmt.Println("\n🤝 It's a draw! The board is full. 🤝")
}