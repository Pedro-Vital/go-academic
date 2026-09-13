// Time to practice what you learned!

// 1) Create a new array (!) that contains three hobbies you have
// 		Output (print) that array in the command line.
// 2) Also output more data about that array:
//		- The first element (standalone)
//		- The second and third element combined as a new list
// 3) Create a slice based on the first element that contains
//		the first and second elements.
//		Create that slice in two different ways (i.e. create two slices in the end)
// 4) Re-slice the slice from (3) and change it to contain the second
//		and last element of the original array.
// 5) Create a "dynamic array" that contains your course goals (at least 2 goals)
// 6) Set the second goal to a different one AND then add a third goal to that existing dynamic array
// 7) Bonus: Create a "Product" struct with title, id, price and create a
//		dynamic list of products (at least 2 products).
//		Then add a third product to the existing list of products.

package main

import "fmt"

type Product struct {
	id    string
	title string
	price float64
}

func main() {
	// 1)
	fmt.Println("Ex 1")
	hobbies := [3]string{"Sports", "Cooking", "Reading"}
	fmt.Println(hobbies)

	// 2)
	fmt.Println("Ex 2")
	fmt.Println(hobbies[0])
	fmt.Println(hobbies[1:3])

	// 3)
	fmt.Println("Ex 3")
	// mainHobbies := hobbies[0:2]
	mainHobbies := hobbies[:2]
	fmt.Println(mainHobbies)
	// A slice is a view of an array.
	// A slice consists of a pointer to an underlying array, a length, and a capacity.

	// 4)
	fmt.Println("Ex 4")
	fmt.Println(cap(mainHobbies))
	mainHobbies = mainHobbies[1:3]
	fmt.Println(mainHobbies)
	// A slice actually just points at a part of an array and we can always select more
	// from that array as long as we expand our selection towards the end of the array.

	// 5)
	fmt.Println("Ex 5")
	courseGoals := []string{"Learn Go!", "Learn all the basics"}
	fmt.Println(courseGoals)
	// The array is created behind the scenes for us

	// 6)
	fmt.Println("Ex 6")
	courseGoals[1] = "Learn all the details!"
	fmt.Println(cap(courseGoals))
	courseGoals = append(courseGoals, "Learn all the basics!")
	fmt.Println(courseGoals)
	fmt.Println(cap(courseGoals))
	// When append doesn't have enough capacity, it allocates a new underlying array. 
	// Reassigning the slice variable makes it point to the new array; the old array 
	// becomes eligible for garbage collection if no other references to it remain.

	// 7)
	fmt.Println("Ex 7")
	products := []Product{
		{
			"first-product",
			"A First Product",
			12.99,
		},
		{
			"second-product",
			"A Second Product",
			129.99,
		},
	}
	// We can omit the struct name because it is already defined as Product.

	fmt.Println(products)

	newProduct := Product{
		"third-product",
		"A Third Product",
		15.99,
	}

	products = append(products, newProduct)

	fmt.Println(products)
}
