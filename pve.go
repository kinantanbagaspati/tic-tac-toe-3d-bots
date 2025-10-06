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
	fmt.Println("2. NaiveMinimaxBot (basic minimax without optimizations)")
	fmt.Println("3. MinimaxBot (optimized minimax with delta evaluation)")
	fmt.Println("4. ConcurrentMinimaxBot (concurrent at top level)")
	fmt.Println("5. ConcurrentMinimaxDeepBot (concurrent at all levels)")
	fmt.Print("Enter your choice (1-5): ")

	var botChoice int
	fmt.Scanln(&botChoice)

	var bot BotInterface
	switch botChoice {
	case 1:
		bot = NewBot('o', "RandomBot")
		fmt.Println("You will face RandomBot!")
	case 2:
		bot = NewNaiveMinimaxBot('o', "NaiveMinimaxBot", 4, 10) // Lower depth for naive approach
		fmt.Println("You will face NaiveMinimaxBot!")
	case 3:
		bot = NewMinimaxBot('o', "MinimaxBot", 6, 10) // Depth 6, Base 10
		fmt.Println("You will face MinimaxBot!")
	case 4:
		bot = NewConcurrentMinimaxBot('o', "ConcurrentMinimaxBot", 6, 10) // Depth 6, Base 10
		fmt.Println("You will face ConcurrentMinimaxBot!")
	case 5:
		bot = NewConcurrentMinimaxDeepBot('o', "ConcurrentMinimaxDeepBot", 5, 10) // Lower depth due to overhead
		fmt.Println("You will face ConcurrentMinimaxDeepBot!")
	default:
		fmt.Println("Invalid choice, defaulting to RandomBot.")
		bot = NewBot('o', "RandomBot")
	}

	totalMoves := 0
	maxMoves := board.Length * board.Width * board.Height

	fmt.Println("\nWelcome to 3D Tic-Tac-Toe!")
	fmt.Printf("You are 'x', %s is 'o'\n", bot.getName())
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
		fmt.Printf("\n%s is thinking...\n", bot.getName())

		start := time.Now()
		botMove, botCoords := bot.MakeMove(board)
		if botCoords[0] == -1 && botCoords[1] == -1 && botCoords[2] == -1 {
			break // No valid moves left
		}
		fmt.Printf("Time taken by %s: %v\n", bot.getName(), time.Since(start))

		fmt.Printf("%s plays %s at coordinates: (%d, %d, %d)\n", bot.getName(), botMove, botCoords[0], botCoords[1], botCoords[2])
		totalMoves++

		// Check for bot win
		winner = board.CheckWin()
		if winner == bot.getSymbol() {
			board.Print()
			fmt.Printf("\n🤖 %s wins! Better luck next time! 🤖\n", bot.getName())
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
