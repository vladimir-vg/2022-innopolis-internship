package main

import "fmt"

func d() {
	go (func() {
		fmt.Println("Deep inside d()")
	})()
}

func f() {
	go d()
}

func g() {
	f()
	go (func() {
		go d()
	})()
}

func main() {
	fmt.Println("Hello World!")
	g()
	d()
}
