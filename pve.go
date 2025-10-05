package main

import (
	"fmt"
	"time"
)

// RunPvE starts a Player vs Environment (Bot) game
func RunPvE() {
	board := NewBoard(3) // Using 3x3x3 for testing purposes

	// Ask user which bot to face
	fmt.Println("🤖 Player vs Bot Mode")
	fmt.Println("Choose your opponent:")
	fmt.Println("1. RandomBot (makes random moves)")
	fmt.Println("2. MinimaxBot (uses strategy)")
	fmt.Println("3. ConcurrentMinimaxBot (uses concurrent strategy)")
	fmt.Print("Enter your choice (1-3): ")

	var botChoice int
	fmt.Scanln(&botChoice)

	var bot BotInterface
	switch botChoice {
	case 1:
		bot = NewBot('o', "RandomBot")
		fmt.Println("You will face RandomBot!")
	case 2:
		bot = NewMinimaxBot('o', "MinimaxBot", 6, 10, 6) // Depth 6, Base 10, Max Power 6
		fmt.Println("You will face MinimaxBot!")
	case 3:
		bot = NewConcurrentMinimaxBot('o', "ConcurrentMinimaxBot", 6, 10, 6) // Depth 6, Base 10, Max Power 6
		fmt.Println("You will face ConcurrentMinimaxBot!")
	default:
		fmt.Println("Invalid choice, defaulting to RandomBot.")
		bot = NewBot('o', "RandomBot")
	}

	totalMoves := 0
	maxMoves := board.Length * board.Width * board.Height

	fmt.Println("\nWelcome to 3D Tic-Tac-Toe!")
	fmt.Printf("You are 'x', %s is 'o'\n", getBotName(bot))
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
		fmt.Printf("\n%s is thinking...\n", getBotName(bot))

		start := time.Now()
		botMove, botCoords := bot.MakeMove(board)
		if botCoords[0] == -1 && botCoords[1] == -1 && botCoords[2] == -1 {
			break // No valid moves left
		}
		fmt.Printf("Time taken by %s: %v\n", getBotName(bot), time.Since(start))

		fmt.Printf("%s plays %s at coordinates: (%d, %d, %d)\n", getBotName(bot), botMove, botCoords[0], botCoords[1], botCoords[2])
		totalMoves++

		// Check for bot win
		winner = board.CheckWin()
		if winner == getBotSymbol(bot) {
			board.Print()
			fmt.Printf("\n🤖 %s wins! Better luck next time! 🤖\n", getBotName(bot))
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

// getBotName returns the name of the bot using type assertion
func getBotName(bot BotInterface) string {
	switch b := bot.(type) {
	case *Bot:
		return b.Name
	case *MinimaxBot:
		return b.Name
	case *ConcurrentMinimaxBot:
		return b.Name
	default:
		return "Unknown Bot"
	}
}

// getBotSymbol returns the symbol of the bot using type assertion
func getBotSymbol(bot BotInterface) byte {
	switch b := bot.(type) {
	case *Bot:
		return b.Symbol
	case *MinimaxBot:
		return b.Symbol
	case *ConcurrentMinimaxBot:
		return b.Symbol
	default:
		return '?'
	}
}
