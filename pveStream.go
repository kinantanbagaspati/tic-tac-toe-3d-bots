package main

import (
	"fmt"
	"strings"
	"time"
)

// RunPvEStream runs the PvE Stream mode with multi-depth concurrent alpha-beta analysis
func RunPvEStream() {
	fmt.Println("🌊 PvE Stream Mode - Multi-Depth Analysis 🌊")
	fmt.Println("═══════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("This mode runs multiple concurrent alpha-beta bots with different depths")
	fmt.Println("and shows real-time analysis as they find better moves!")
	fmt.Println()

	// Create a new board
	board := NewBoard()

	// Player is always X, multi-depth bot is O
	playerSymbol := byte('x')
	botSymbol := byte('o')
	currentPlayer := playerSymbol

	// Define the depths to analyze
	depths := []int{3, 4, 5, 6, 7}

	fmt.Printf("Analyzing with depths: %v\n", depths)
	fmt.Println()

	for {
		board.Print()
		fmt.Println()

		// Check for win condition
		winner := board.CheckWin()
		if winner != '|' {
			if winner == playerSymbol {
				fmt.Println("🎉 You win! 🎉")
			} else {
				fmt.Println("🤖 Bot wins! 🤖")
			}
			break
		}

		// Check for draw
		if len(board.GetValidMoves()) == 0 {
			fmt.Println("🤝 It's a draw! 🤝")
			break
		}

		if currentPlayer == playerSymbol {
			// Player's turn
			fmt.Print("Your turn! Enter move (e.g., A1, B2, C3): ")
			var moveInput string
			fmt.Scanln(&moveInput)

			col, row := parseMove(moveInput)
			if col == -1 || row == -1 {
				fmt.Println("Invalid format! Use format like A1, B2, C3")
				continue
			}

			coords := board.Move(moveInput, playerSymbol)
			if coords[0] == -1 {
				fmt.Println("Invalid move! Try again.")
				continue
			}

			fmt.Printf("You played %s at (%d, %d, %d)\n", moveInput, coords[0], coords[1], coords[2])
		} else {
			// Multi-depth bot's turn
			fmt.Println("🤖 Multi-Depth Bot is analyzing...")
			fmt.Println("─────────────────────────────────────")

			start := time.Now()

			// Use multi-depth streaming analysis
			resultCh := multiDepthAlphaBetaStream(board, false, depths) // Bot is minimizing (O)

			var bestMove string
			var finalResult MultiDepthStreamResult

			// Listen to the stream and show real-time updates
			for result := range resultCh {
				if result.Final {
					finalResult = result
					break
				}

				// Show intermediate results
				movesStr := strings.Join(result.Moves, " → ")
				fmt.Printf("📈 New best move from depth %d: [%s] (Score: %d)\n",
					result.Depth, movesStr, result.Score)
			}

			duration := time.Since(start)

			// Execute the best move found
			if len(finalResult.Moves) > 0 {
				bestMove = finalResult.Moves[0]
				coords := board.Move(bestMove, botSymbol)

				fmt.Println("─────────────────────────────────────")
				movesStr := strings.Join(finalResult.Moves, " → ")
				fmt.Printf("🎯 Final decision from depth %d: [%s]\n", finalResult.Depth, movesStr)
				fmt.Printf("🤖 Bot plays %s at (%d, %d, %d) - Time: %v\n",
					bestMove, coords[0], coords[1], coords[2], duration)
			} else {
				fmt.Println("🤖 Bot cannot find a valid move!")
				break
			}
		}

		fmt.Println()

		// Switch players
		if currentPlayer == playerSymbol {
			currentPlayer = botSymbol
		} else {
			currentPlayer = playerSymbol
		}
	}

	fmt.Println("\nGame Over! Thanks for playing! 👋")
}
