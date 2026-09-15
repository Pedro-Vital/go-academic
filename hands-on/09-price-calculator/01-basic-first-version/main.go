package main

import "fmt"

func main() {
	prices := []float64{10, 20, 30}
	taxRates := []float64{0, 0.07, 0.1, 0.15}

	result := make(map[float64][]float64)

	for _, taxRate := range taxRates {
		taxIncludedPrices := make([]float64, len(prices))
		for priceIndex, price := range prices {
			taxIncludedPrices[priceIndex] = price * (1 + taxRate)
		}
		result[taxRate] = taxIncludedPrices
		// For every tax rate, I have the corresponding prices
	}

	fmt.Println(result)
	// map[0:[10 20 30] 
	// 	   0.07:[10.700000000000001 21.400000000000002 32.1] 
	//     0.1:[11 22 33] 0.15:[11.5 23 34.5]]
}