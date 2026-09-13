// By default, Go creates a copy when passing values to functions
// For very large and complex values, this may take up too much memory space
// With pointers, only one value is stored in memory

// Here we are passing the pointer to a function to avoid unnecessary copies

package main

import "fmt"

func main() {
	age := 32

	var agePointer *int // the asterisk tells it's a pointer

	agePointer = &age

	fmt.Println("Pointer:", agePointer)

	fmt.Println("Age:", *agePointer) // this is called dereferencing

	// We use an ampersand in front of a normal value to get its pointer
	// and we use an asterisk in front of a pointer to get its value

	adultYears := getAdultYears(agePointer)
	fmt.Println(adultYears)
}

func getAdultYears(age *int) int {
	return *age - 18
}
