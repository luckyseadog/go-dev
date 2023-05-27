package main

import (
	"github.com/luckyseadog/go-dev/internal/agent"
	"time"
)

func main() {
	agent := agent.NewAgent(":8080", "text/plain", 2*time.Second, 10*time.Second)
	agent.Run()
	time.Sleep(20 * time.Second)
}
