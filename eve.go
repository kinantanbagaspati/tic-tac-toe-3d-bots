package main

import (
	"fmt"
	"time"
)

// BotStats tracks performance statistics for a bot
type BotStats struct {
	Name        string
	TotalTime   time.Duration
	MoveCount   int
	AverageTime time.Duration
}

// UpdateStats adds a move time to the bot's statistics
func (stats *BotStats) UpdateStats(moveTime time.Duration) {
	stats.TotalTime += moveTime
	stats.MoveCount++
	stats.AverageTime = stats.TotalTime / time.Duration(stats.MoveCount)
}

// RunEvE starts an Environment vs Environment (Bot vs Bot) game
func RunEvE() {
	board := NewBoard(3) // Using 3x3x3 for testing purposes

	fmt.Println("🤖 Bot vs Bot Mode (Eve) 🤖")
	fmt.Println("Choose the bots to fight:")

	// Select first bot (X player)
	fmt.Println("\nSelect Bot 1 (plays 'x'):")
	fmt.Println("1. RandomBot (makes random moves)")
	fmt.Println("2. MinimaxBot (uses strategy)")
	fmt.Println("3. ConcurrentMinimaxBot (uses concurrent strategy)")
	fmt.Print("Enter your choice (1-3): ")

	var bot1Choice int
	fmt.Scanln(&bot1Choice)

	bot1 := createBot(bot1Choice, 'x', "Bot1")
	if bot1 == nil {
		fmt.Println("Invalid choice, defaulting to RandomBot.")
		bot1 = NewBot('x', "RandomBot")
	}

	// Select second bot (O player)
	fmt.Println("\nSelect Bot 2 (plays 'o'):")
	fmt.Println("1. RandomBot (makes random moves)")
	fmt.Println("2. MinimaxBot (uses strategy)")
	fmt.Println("3. ConcurrentMinimaxBot (uses concurrent strategy)")
	fmt.Print("Enter your choice (1-3): ")

	var bot2Choice int
	fmt.Scanln(&bot2Choice)

	bot2 := createBot(bot2Choice, 'o', "Bot2")
	if bot2 == nil {
		fmt.Println("Invalid choice, defaulting to RandomBot.")
		bot2 = NewBot('o', "RandomBot")
	}

	// Initialize statistics
	bot1Stats := &BotStats{Name: bot1.getName()}
	bot2Stats := &BotStats{Name: bot2.getName()}

	totalMoves := 0
	maxMoves := board.Length * board.Width * board.Height

	fmt.Println("\n🎯 Bot Battle Begins! 🎯")
	fmt.Printf("%s ('x') vs %s ('o')\n", bot1Stats.Name, bot2Stats.Name)
	fmt.Println("Press Enter to continue between moves, or type 'auto' for automatic play...")

	var playMode string
	fmt.Scanln(&playMode)
	autoPlay := playMode == "auto"

	for totalMoves < maxMoves {
		if !autoPlay {
			board.Print()
		}

		// Bot 1's turn (X)
		fmt.Printf("\n%s ('x') is thinking...\n", bot1Stats.Name)

		start := time.Now()
		bot1Move, bot1Coords := bot1.MakeMove(board)
		moveTime := time.Since(start)
		bot1Stats.UpdateStats(moveTime)

		if bot1Coords[0] == -1 && bot1Coords[1] == -1 && bot1Coords[2] == -1 {
			break // No valid moves left
		}

		fmt.Printf("%s plays %s at (%d, %d, %d) - Time: %v (Avg: %v)\n",
			bot1Stats.Name, bot1Move, bot1Coords[0], bot1Coords[1], bot1Coords[2],
			moveTime, bot1Stats.AverageTime)
		totalMoves++

		// Check for bot1 win
		winner := board.CheckWin()
		if winner == 'x' {
			if !autoPlay {
				board.Print()
			}
			fmt.Printf("\n🎉 %s ('x') wins! 🎉\n", bot1Stats.Name)
			printFinalStats(bot1Stats, bot2Stats)
			return
		}

		// Check if board is full
		if board.IsFull() {
			break
		}

		if !autoPlay {
			fmt.Print("Press Enter to continue...")
			fmt.Scanln()
		}

		// Bot 2's turn (O)
		fmt.Printf("\n%s ('o') is thinking...\n", bot2Stats.Name)

		start = time.Now()
		bot2Move, bot2Coords := bot2.MakeMove(board)
		moveTime = time.Since(start)
		bot2Stats.UpdateStats(moveTime)

		if bot2Coords[0] == -1 && bot2Coords[1] == -1 && bot2Coords[2] == -1 {
			break // No valid moves left
		}

		fmt.Printf("%s plays %s at (%d, %d, %d) - Time: %v (Avg: %v)\n",
			bot2Stats.Name, bot2Move, bot2Coords[0], bot2Coords[1], bot2Coords[2],
			moveTime, bot2Stats.AverageTime)
		totalMoves++

		// Check for bot2 win
		winner = board.CheckWin()
		if winner == 'o' {
			if !autoPlay {
				board.Print()
			}
			fmt.Printf("\n🎉 %s ('o') wins! 🎉\n", bot2Stats.Name)
			printFinalStats(bot1Stats, bot2Stats)
			return
		}

		// Check if board is full
		if board.IsFull() {
			break
		}

		if !autoPlay {
			fmt.Print("Press Enter to continue...")
			fmt.Scanln()
		}
	}

	// If we reach here, it's a draw
	if !autoPlay {
		board.Print()
	}
	fmt.Println("\n🤝 It's a draw! The board is full. 🤝")
	printFinalStats(bot1Stats, bot2Stats)
}

// createBot creates a bot based on user choice
func createBot(choice int, symbol byte, defaultName string) BotInterface {
	switch choice {
	case 1:
		return NewBot(symbol, defaultName)
	case 2:
		return NewMinimaxBot(symbol, defaultName, 6, 10, 6)
	case 3:
		return NewConcurrentMinimaxBot(symbol, defaultName, 6, 10, 6)
	default:
		return nil
	}
}

// printFinalStats displays the final performance statistics
func printFinalStats(bot1Stats, bot2Stats *BotStats) {
	fmt.Println("\n📊 Final Performance Statistics 📊")
	fmt.Println("═══════════════════════════════════════")

	fmt.Printf("🤖 %s:\n", bot1Stats.Name)
	fmt.Printf("   Total Moves: %d\n", bot1Stats.MoveCount)
	fmt.Printf("   Total Time:  %v\n", bot1Stats.TotalTime)
	fmt.Printf("   Average Time: %v\n", bot1Stats.AverageTime)

	fmt.Printf("\n🤖 %s:\n", bot2Stats.Name)
	fmt.Printf("   Total Moves: %d\n", bot2Stats.MoveCount)
	fmt.Printf("   Total Time:  %v\n", bot2Stats.TotalTime)
	fmt.Printf("   Average Time: %v\n", bot2Stats.AverageTime)

	// Performance comparison
	fmt.Println("\n⚡ Performance Comparison:")
	if bot1Stats.AverageTime < bot2Stats.AverageTime {
		ratio := float64(bot2Stats.AverageTime) / float64(bot1Stats.AverageTime)
		fmt.Printf("   %s is %.2fx faster than %s\n", bot1Stats.Name, ratio, bot2Stats.Name)
	} else if bot2Stats.AverageTime < bot1Stats.AverageTime {
		ratio := float64(bot1Stats.AverageTime) / float64(bot2Stats.AverageTime)
		fmt.Printf("   %s is %.2fx faster than %s\n", bot2Stats.Name, ratio, bot1Stats.Name)
	} else {
		fmt.Println("   Both bots have similar performance!")
	}
}
