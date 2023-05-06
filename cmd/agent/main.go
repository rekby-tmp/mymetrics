package main

import "time"

func main() {
	agent := NewAgent("http://localhost:8080", time.Second*2, time.Second*10)
	agent.Start()
}
