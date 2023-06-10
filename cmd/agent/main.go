package main

import (
	"github.com/luckyseadog/go-dev/internal/agent"
	"time"
)

func main() {
	agent := agent.NewAgent("http://127.0.0.1:8080", "application/json", 2*time.Second, 10*time.Second)
	time.AfterFunc(2*time.Minute, func() {
		agent.Stop()
	})
	agent.Run()
}
