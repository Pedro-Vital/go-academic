package main

import "fmt"

func main() {

	var x [100]int // array

	s := x[:10] // slice
	fmt.Printf("x type: %T | s type: %T\n", x, s)
	fmt.Println("s length:", len(s))

	y := []int{} // slice
	// This is not an allocated array with zero values.
	// We can't do
	// y[0] = 3
	fmt.Printf("y type: %T\n", y)
	fmt.Println("y length:", len(y))
	fmt.Println("y capacity:", cap(y))

	for i := 1; i <= 50; i++ {
		y = append(y, i)
		fmt.Printf("Len: %v\n",len(y))
		fmt.Printf("Cap: %v\n",cap(y))
		// When append doesn't have enough capacity, it allocates a new underlying array. 
		// Reassigning the slice variable makes it point to the new array; the old array 
		// becomes eligible for garbage collection if no other references to it remain.
	}

	fmt.Println(y)
}
