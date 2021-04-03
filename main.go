package main

import (
	"fmt"
	"time"
)

func main() {
	go hello()
	time.Sleep(1 * time.Second)
	fmt.Println("main function")
}

func hello()  {
	fmt.Println("Hello goroutine")
}