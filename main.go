package main

import "fmt"

func main() {
	c := make(chan int)
	go producer(c)
	for value := range c{
		fmt.Println("value: ", value)
	}
	fmt.Println("Channel has been closed!")
}

func producer(c chan int)  {
	for i := 0; i <= 10; i++{
		c <- i
	}
	close(c)
}

