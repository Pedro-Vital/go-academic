package fileops

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Uppercase first letter to export identifier
func GetFloatFromFile(fileName string) (float64, error) {
	data, err := os.ReadFile(fileName)

	if err != nil { // nil stands for zero values, similar to null or none. When err == nil, the operation succeeded
		return 0, errors.New("Failed to find file.")
	}

	// Suggested way to get the value from the file as float64:
	valueText := string(data)
	value, err := strconv.ParseFloat(valueText, 64)

	if err != nil {
		return 0, errors.New("Failed to parse stored value.")
	}

	return value, nil
}

// Uppercase first letter to export identifier
func WriteFloatToFile(value float64, fileName string) {
	valueText := fmt.Sprint(value)
	os.WriteFile(fileName, []byte(valueText), 0644)
	// the string is converted to such a byte collection using []bite()
	// 0644 is a file permission encoding. For more: https://www.redhat.com/en/blog/linux-file-permissions-explained
}
