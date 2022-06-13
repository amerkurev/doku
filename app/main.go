package main

import "fmt"

var revision = "unknown"

func dummy(a, b int) int {
	return a + b
}

func main() {
	fmt.Printf("doku %s\n", revision)

	fmt.Println("Shutting down")
}
