package main

func beta() {
	go (func() {})()
	// 6
	go alpha()
}
