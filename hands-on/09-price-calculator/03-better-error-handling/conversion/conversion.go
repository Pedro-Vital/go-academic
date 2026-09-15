package conversion

import (
	"errors"
	"strconv"
)

// I  wanted to outsource the conversion logic of DataLoad().
// That is the purpose of this package.

func StringsToFloats(strings []string) ([]float64, error) {
	var floats []float64

	for _, stringVal := range strings {
		floatVal, err := strconv.ParseFloat(stringVal, 64)

		if err != nil {
			return nil, errors.New("Failed to convert string to float")
		}

		floats = append(floats, floatVal)
	}

	return floats, nil
}
