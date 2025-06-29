package main

import (
	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/logger"
	"neuroshell/internal/shell"
)

func main() {
	logger.Info("Starting NeuroShell v0.1.0")
	
	// Initialize services before starting shell
	if err := shell.InitializeServices(); err != nil {
		logger.Fatal("Failed to initialize services", "error", err)
	}
	
	logger.Info("Services initialized successfully")

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