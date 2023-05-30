package main

import "time"

func main() {
	agent := NewAgent("http://127.0.0.1:8080", "text/plain", 2*time.Second, 10*time.Second)
	time.AfterFunc(30*time.Second, func() {
		agent.Stop()
	})
	agent.Run()
}
