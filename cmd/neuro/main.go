package main

import (
	"github.com/abiosoft/ishell/v2"
	"neuroshell/internal/shell"
)

func main() {
	sh := ishell.New()
	
	sh.Println("Neuro Shell v0.1.0 - LLM-integrated shell environment")
	sh.Println("Type '\\help' for Neuro commands or 'exit' to quit.")
	
	sh.NotFound(shell.ProcessInput)
	
	sh.Run()
}