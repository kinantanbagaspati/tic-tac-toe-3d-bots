package main

import "fmt"

func main() {
	fmt.Println("🎯 Welcome to 3D Tic-Tac-Toe! 🎯")
	fmt.Println("═══════════════════════════════")
	fmt.Println()
	fmt.Println("Choose game mode:")
	fmt.Println("1. Player vs Player (PvP)")
	fmt.Println("2. Player vs Bot (PvE)")
	fmt.Println("3. Bot vs Bot (Eve)")
	fmt.Println("4. PvE Stream (Multi-Depth Analysis)")
	fmt.Println("5. EvE Stream (Bidirectional Persistent Search)")
	fmt.Println("6. Exit")
	fmt.Println()

	var choice int
	fmt.Print("Enter your choice (1-6): ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		RunPvP()
	case 2:
		RunPvE()
	case 3:
		RunEvE()
	case 4:
		RunPvEStream()
	case 5:
		RunEvEStream()
	case 6:
		fmt.Println("Thanks for playing! Goodbye! 👋")
	default:
		fmt.Println("Invalid choice. Please run the program again and select 1, 2, 3, 4, 5, or 6.")
	}
}
