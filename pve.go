package main

import (
	"fmt"
	"time"
)

// RunPvE starts a Player vs Environment (Bot) game
func RunPvE() {
	board := NewBoard(3) // Using 3x3x3 for testing purposes
	bot := NewBot('o', "RandomBot")
	
	// Alternative bot options (uncomment to use):
	// bot := NewMinimaxBot('o', "MinimaxBot", 3) // Depth 3 search
	
	totalMoves := 0
	maxMoves := board.Length * board.Width * board.Height
	
	fmt.Println("🤖 Player vs Bot Mode")
	fmt.Println("Welcome to 3D Tic-Tac-Toe!")
	fmt.Printf("You are 'x', Bot is '%c'\n", bot.Symbol)
	fmt.Printf("Enter moves in format like A1, B2, etc. (A-%c, 1-%d)\n", 'A'+byte(board.Length-1), board.Width)
	fmt.Println()
	
	for totalMoves < maxMoves {
		board.Print()
		
		// Player's turn
		fmt.Printf("\nYour turn (playing 'x'): ")
		var moveInput string
		fmt.Scanln(&moveInput)
		
		coords := board.Move(moveInput, 'x')
		if coords[0] == -1 && coords[1] == -1 && coords[2] == -1 {
			fmt.Println("Invalid move! Try again.")
			continue
		}
		
		fmt.Printf("Your move %s placed at coordinates: (%d, %d, %d)\n", moveInput, coords[0], coords[1], coords[2])
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
		
		botMove, botCoords := bot.MakeMove(board)
		if botCoords[0] == -1 && botCoords[1] == -1 && botCoords[2] == -1 {
			break // No valid moves left
		}
		
		fmt.Printf("%s plays %s at coordinates: (%d, %d, %d)\n", bot.Name, botMove, botCoords[0], botCoords[1], botCoords[2])
		totalMoves++
		
		// Check for bot win
		winner = board.CheckWin()
		if winner == bot.Symbol {
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
