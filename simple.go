package main

import "fmt"

type Hello struct {
	A int
}

func main() {
	a := 2
	a2 := 4 + a
	fmt.Println(a)
	fmt.Println(a2)
	fmt.Println("Hello int world!")
}
