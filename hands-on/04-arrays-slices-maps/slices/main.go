package main

import "fmt"

func main() {

	var x [100]int // array

	s := x[:10] // slice
	fmt.Printf("x type: %T | s type: %T\n", x, s)
	fmt.Println("s length:", len(s))

	y := []int{} // slice
	fmt.Printf("y type: %T\n", y)
	fmt.Println("y length:", len(y))

	for i := 1; i <= 200; i++ {
		y = append(y, i)
		fmt.Println(len(y))
	}

	fmt.Println(y)
}
