package main

import (
	"time"
)

func main() {
	agent := NewAgent(":8080", "text/plain", 2*time.Second, 10*time.Second)
	agent.Run()
	time.Sleep(5 * time.Minute)
}
