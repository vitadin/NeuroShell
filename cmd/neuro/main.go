package main

import (
	"log"

	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/shell"
)

func main() {
	// Initialize services before starting shell
	if err := shell.InitializeServices(); err != nil {
		log.Fatal("Failed to initialize services:", err)
	}

	sh := ishell.New()
	sh.SetPrompt("neuro> ")
	
	// Remove built-in commands so they become user messages or Neuro commands
	sh.DeleteCmd("exit")
	sh.DeleteCmd("help")
	
	sh.Println("Neuro Shell v0.1.0 - LLM-integrated shell environment")
	sh.Println("Type '\\help' for Neuro commands or '\\exit' to quit.")
	
	sh.NotFound(shell.ProcessInput)
	
	sh.Run()
}