package main

import "fmt"

func main() {
	fmt.Println("Hello World!")
	// 1
	go (func() {})()
	// 2
	go (func() {
		// 5
		go (func() {})()
		// 6
		go alpha()
	})()
	// 3
	go alpha()
	//
	go beta()
	// 4
	go alpha()
}
