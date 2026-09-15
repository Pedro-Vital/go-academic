package main

// In the main function, I want to generate a bunch of jobs, where every job
// should be based on a single tax rate. Then I want to dispatch those jobs.
// Inside those jobs, prices should be read from a file and the calculation
// of new prices based on the tax value should be performed.
// Later, during the job execution, the results must be written to a file.

import (
	"fmt"
	"price-calculator/filemanager"
	"price-calculator/prices"
)

func main() {
	taxRates := []float64{0, 0.07, 0.1, 0.15}
	doneChans := make([]chan bool, len(taxRates))
	errorChans := make([]chan error, len(taxRates))

	// Now for every tax rate I want to create a job:
	for index, taxRate := range taxRates {
		doneChans[index] = make(chan bool)
		errorChans[index] = make(chan error)

		fm := filemanager.New("prices.txt", fmt.Sprintf("result_%.0f.json", taxRate*100))
		// cmdm := cmdmanager.New()
		// We can easily switch the io mechanism between the FileManager and other 
		// structs that implement the iomanager interface (e.g. the CMDManager). 
		priceJob := prices.NewTaxIncludedPricesJob(fm, taxRate)
		
		go priceJob.Process(doneChans[index], errorChans[index])

		// if err != nil {
		// 	fmt.Println("Could not process job")
		// 	fmt.Println(err)
		// }
	}

	// It could be like that if we weren't handling errors
	// for _,doneChan := range doneChans {
	// 	<-doneChan
	// }

	// We need a for loop that goes through all channels that we could have
	for index := range taxRates {
		select {
		case err := <-errorChans[index]:
			if err != nil {
				fmt.Println(err)
			}
		case <-doneChans[index]:
			fmt.Println("Done!")
		}
	}

}
