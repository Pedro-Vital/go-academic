package main

import "fmt"

func main() {

	ages := make(map[string]int) // declaration with make
	// We can let Go know the expected length of the array or map for efficiency. 
	// To do that, we just pass the size argument for the make() function.

	ages["Pedro"] = 25
	ages["Teo"] = 33

	fmt.Println("Ages:", ages)

	heights := map[string]float64{} // most common declaration
	fmt.Println("Len:", len(heights))
	// Unlike slices,
	// We can do:
	heights["Pedro"] = 1.76
	// for that type of declaration
	fmt.Println("Heights:", heights)

	heights["Teo"] = 1.82
	heightTeo, ok := heights["Teo"]
	if ok {
		fmt.Println("Height Teo:", heightTeo, "ok:", ok)
	} else {
		fmt.Println("I didn't find it.")
	}

	// We can do it like this:
	if heightTeo, ok := heights["Teo"]; ok {
		fmt.Println("Height Teo:", heightTeo)
	} else {
		fmt.Println("I didn't find it.")
	}
	// Here, heightTeo and ok exists only within the conditional structure

}
