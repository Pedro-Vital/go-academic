package main

import (
	"fmt"
	"math"
)

const inflationRate = 2.5 // constant

func main() {
	expectedReturnRate := 4.3 // inferred type float64

	// var investmentAmount, years float64 = 1000, 10
	// Let's use input from the user instead
	var investmentAmount, years float64

	fmt.Println("Enter the investment amount:")
	fmt.Scanln(&investmentAmount)
	fmt.Println("Enter the number of years:")
	fmt.Scanf("%f", &years) // specifying the format of the input

	futureValue, realFutureValue := calculateFutureValue(investmentAmount, expectedReturnRate, years)
	//futureValue := investmentAmount * math.Pow(1 + expectedReturnRate/100, years)
	//realFutureValue := futureValue / math.Pow(1 + inflationRate/100, years)

	// 1. Option: Print the output information using fmt.Printf
	// fmt.Printf("Future Value: %.2f\nFuture Value (adjusted for Inflation): %.2f", futureValue, realFutureValue)

	// 2. Option: Print the output information using fmt.Sprintf
	// The fmt.Sprintf function formats the string and returns it as a string
	formattedFV := fmt.Sprintf("Future Value: %.2f\n", futureValue)
	formattedRFV := fmt.Sprintf("Future Value (adjusted for Inflation): %.2f\n", realFutureValue)

	fmt.Print(formattedFV, formattedRFV)
}

func calculateFutureValue(investmentAmount, expectedReturnRate, years float64) (fv float64, rfv float64) /*(float64, float64) works too*/ {
	fv = investmentAmount * math.Pow(1+expectedReturnRate/100, years)
	rfv = fv / math.Pow(1+inflationRate/100, years)
	return fv, rfv
	// return
	// we can also use just return and it will return the values defined in the first line of the function as output
}
