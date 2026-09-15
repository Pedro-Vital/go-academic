package prices

import (
	"fmt"

	"price-calculator/conversion"
	"price-calculator/iomanager"
)

type TaxIncludedPriceJob struct {
	IOManager         iomanager.IOManager `json:"-"` // "-" means "ignore"
	TaxRate           float64             `json:"tax_rate"`
	InputPrices       []float64           `json:"input_prices"`
	TaxIncludedPrices map[string]string   `json:"tax_included_prices"`
	// We could simply use []float64 for this latter one, but let's do it with a map.
	// The key (string) will be later the input price again, for which we have a tax included price.
}

func (job *TaxIncludedPriceJob) LoadData() {

	lines, err := job.IOManager.ReadLines()

	if err != nil {
		fmt.Println(err)
		return
	}

	prices, err := conversion.StringsToFloats(lines)

	if err != nil {
		fmt.Println(err)
		return
	}

	job.InputPrices = prices
	// We are mutating the instance data, so it must be a pointer receiver.
}

// Here we have calculation for the tax included price.
func (job *TaxIncludedPriceJob) Process() {
	job.LoadData()
	// LoadData has a pointer receiver because it modifies the job.
	// Process also uses a pointer receiver so that LoadData operates on
	// the same job instance rather than on a copy.
	// If we call job.LoadData() with Process being a value receiver,
	// the pointer received by LoadData would be the copy pointer!!!

	result := make(map[string]string)
	for _, price := range job.InputPrices {
		taxIncludedPrice := price * (1 + job.TaxRate)
		result[fmt.Sprintf("%.2f", price)] = fmt.Sprintf("%.2f", taxIncludedPrice)
	}

	job.TaxIncludedPrices = result // another reason it should be a pointer receiver
	job.IOManager.WriteResult(job)

	// fmt.Println(result)

	/*
		InputPrices: []float64{10, 20, 30}

		map[10.00:10 20.00:20 30.00:30]
		map[10.00:10.70 20.00:21.40 30.00:32.1]
		map[10.00:11 20.00:22 30.00:33]
		map[10.00:11.5 20.00:23 30.00:34.5]
	*/
}

// Constructor
// If we were in a price package and we had a price struct, so to say,
// we could simply name the constructor "New()". But that's not the case here.
func NewTaxIncludedPricesJob(iom iomanager.IOManager, taxRate float64) *TaxIncludedPriceJob {
	return &TaxIncludedPriceJob{
		IOManager:   iom,
		InputPrices: []float64{10, 20, 30},
		TaxRate:     taxRate,
		// TaxIncludedPrices will be filled with calculations
	}
}
