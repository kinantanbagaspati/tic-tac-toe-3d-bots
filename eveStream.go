package main

import (
	"fmt"
	"time"
)

// RunEvEStream runs the EvE Stream mode where two persistent minimax bots face each other
// with bidirectional streaming and background calculation during opponent thinking time
func RunEvEStream() {
	fmt.Println("🤖⚔️🤖 EvE Stream Mode - Bidirectional Persistent Search 🤖⚔️🤖")
	fmt.Println("═════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("Two persistent minimax bots with background calculation!")
	fmt.Println("Each bot continues calculating during opponent's thinking time.")
	fmt.Println()

	// Create a new board
	board := NewBoard()

	// Create two persistent minimax bots
	botX := NewPersistentMinimaxBot('x', "PersistentBot-X", 4, 10)
	botO := NewPersistentMinimaxBot('o', "PersistentBot-O", 4, 10)

	// Ensure cleanup at the end
	defer botX.Close()
	defer botO.Close()

	currentPlayer := byte('x')
	moveCount := 0

	fmt.Printf("🤖 %s (X) vs %s (O) 🤖\n", botX.getName(), botO.getName())
	fmt.Println()

	for {
		board.Print()
		fmt.Println()

		// Check for win condition
		winner := board.CheckWin()
		if winner != '|' {
			if winner == 'x' {
				fmt.Printf("🎉 %s (X) wins! 🎉\n", botX.getName())
			} else {
				fmt.Printf("🎉 %s (O) wins! 🎉\n", botO.getName())
			}
			break
		}

		// Check for draw
		if len(board.GetValidMoves()) == 0 {
			fmt.Println("🤝 It's a draw! 🤝")
			break
		}

		moveCount++
		fmt.Printf("Move %d: ", moveCount)

		var move string
		var coords [3]int
		var activeBot *PersistentMinimaxBot
		var waitingBot *PersistentMinimaxBot

		if currentPlayer == 'x' {
			activeBot = botX
			waitingBot = botO
			fmt.Printf("%s (X) is thinking...", botX.getName())
		} else {
			activeBot = botO
			waitingBot = botX
			fmt.Printf("%s (O) is thinking...", botO.getName())
		}

		// Measure thinking time
		start := time.Now()

		// Active bot makes a move (this triggers background calculation in waiting bot)
		move, coords = activeBot.MakeMove(board)

		duration := time.Since(start)

		if coords[0] == -1 {
			fmt.Printf("\n🚨 %s cannot find a valid move!\n", activeBot.getName())
			break
		}

		fmt.Printf(" -> %s at (%d, %d, %d) [Time: %v]\n",
			move, coords[0], coords[1], coords[2], duration)

		// Notify the waiting bot about opponent's move for tree pruning
		waitingBot.OpponentMove(move)

		// Show some statistics about the bots' search trees
		showSearchStats(activeBot, waitingBot, duration)

		fmt.Println()

		// Switch players
		if currentPlayer == 'x' {
			currentPlayer = 'o'
		} else {
			currentPlayer = 'x'
		}

		// Small delay to make it more watchable
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\nGame Over! Both bots performed persistent background search! 🎯")

	// Show final statistics
	fmt.Println("\n📊 Final Search Statistics:")
	showFinalStats(botX, botO)
}

// showSearchStats displays current search statistics for both bots
func showSearchStats(activeBot, waitingBot *PersistentMinimaxBot, thinkingTime time.Duration) {
	fmt.Printf("   📈 Search Stats - Active: %s, Background: %s\n",
		activeBot.getName(), waitingBot.getName())

	// Get node counts (simplified for now)
	activeNodes := getNodeCount(activeBot)
	waitingNodes := getNodeCount(waitingBot)

	fmt.Printf("   🔍 Active bot nodes: %d, Background bot nodes: %d\n",
		activeNodes, waitingNodes)

	fmt.Printf("   ⏱️  Thinking time: %v (Background bot was calculating simultaneously)\n",
		thinkingTime)
}

// showFinalStats displays final statistics for both bots
func showFinalStats(botX, botO *PersistentMinimaxBot) {
	fmt.Printf("🤖 %s final nodes: %d\n", botX.getName(), getNodeCount(botX))
	fmt.Printf("🤖 %s final nodes: %d\n", botO.getName(), getNodeCount(botO))
	fmt.Println("Both bots maintained persistent search trees throughout the game!")
}

// getNodeCount returns the number of nodes in a bot's search tree
func getNodeCount(bot *PersistentMinimaxBot) int {
	if bot.tree == nil {
		return 0
	}

	bot.tree.mutex.RLock()
	count := len(bot.tree.nodes)
	bot.tree.mutex.RUnlock()

	return count
}
