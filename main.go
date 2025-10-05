package main

import "fmt"

func main() {
	fmt.Println("🎯 Welcome to 3D Tic-Tac-Toe! 🎯")
	fmt.Println("═══════════════════════════════")
	fmt.Println()
	fmt.Println("Choose game mode:")
	fmt.Println("1. Player vs Player (PvP)")
	fmt.Println("2. Player vs Bot (PvE)")
	fmt.Println("3. Exit")
	fmt.Println()
	
	var choice int
	fmt.Print("Enter your choice (1-3): ")
	fmt.Scanln(&choice)
	
	switch choice {
	case 1:
		RunPvP()
	case 2:
		RunPvE()
	case 3:
		fmt.Println("Thanks for playing! Goodbye! 👋")
	default:
		fmt.Println("Invalid choice. Please run the program again and select 1, 2, or 3.")
	}
}
