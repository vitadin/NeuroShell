package main

import (
	"github.com/abiosoft/ishell/v2"
)

func main() {
	shell := ishell.New()
	
	shell.Println("Neuro Shell v0.1.0 - LLM-integrated shell environment")
	shell.Println("Type 'help' for available commands or 'exit' to quit.")
	
	shell.AddCmd(&ishell.Cmd{
		Name: "send",
		Help: "send message to LLM agent",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: \\send message")
				return
			}
			message := c.Args[0]
			c.Printf("Sending: %s\n", message)
		},
	})
	
	shell.AddCmd(&ishell.Cmd{
		Name: "set",
		Help: "set a variable",
		Func: func(c *ishell.Context) {
			if len(c.Args) < 2 {
				c.Println("Usage: \\set variable value")
				return
			}
			variable := c.Args[0]
			value := c.Args[1]
			c.Printf("Setting %s = %s\n", variable, value)
		},
	})
	
	shell.AddCmd(&ishell.Cmd{
		Name: "get",
		Help: "get a variable",
		Func: func(c *ishell.Context) {
			if len(c.Args) == 0 {
				c.Println("Usage: \\get variable")
				return
			}
			variable := c.Args[0]
			c.Printf("Getting %s\n", variable)
		},
	})

	shell.Run()
}