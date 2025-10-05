package main

import "fmt"

// RunPvP starts a Player vs Player game
func RunPvP() {
	board := NewBoard(3) // Using 3x3x3 for testing purposes
	
	players := []byte{'x', 'o'}
	playerNames := []string{"Player X", "Player O"}
	currentPlayer := 0
	totalMoves := 0
	maxMoves := board.Length * board.Width * board.Height
	
	fmt.Println("🎮 Player vs Player Mode")
	fmt.Println("Welcome to 3D Tic-Tac-Toe!")
	fmt.Printf("Enter moves in format like A1, B2, etc. (A-%c, 1-%d)\n", 'A'+byte(board.Length-1), board.Width)
	fmt.Println()
	
	for totalMoves < maxMoves {
		board.Print()
		fmt.Printf("\n%s's turn (playing '%c'): ", playerNames[currentPlayer], players[currentPlayer])
		
		var moveInput string
		fmt.Scanln(&moveInput)
		
		coords := board.Move(moveInput, players[currentPlayer])

		if coords[0] == -1 && coords[1] == -1 && coords[2] == -1 {
			fmt.Println("Invalid move! Try again.")
			continue
		}
		
		fmt.Printf("Move %s placed at coordinates: (%d, %d, %d)\n", moveInput, coords[0], coords[1], coords[2])
		totalMoves++
		
		// Check for win
		winner := board.CheckWin()
		if winner != '|' {
			board.Print()
			fmt.Printf("\n🎉 %s wins! 🎉\n", playerNames[currentPlayer])
			return
		}
		
		// Switch to next player
		currentPlayer = (currentPlayer + 1) % 2
	}
	
	// If we reach here, it's a draw
	board.Print()
	fmt.Println("\n🤝 It's a draw! The board is full. 🤝")
}
