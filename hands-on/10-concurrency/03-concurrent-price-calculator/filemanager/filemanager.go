package filemanager

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"time"
)

// I could simply use ReadLines() and WriteResult() right away in prices,
// but with the FileManager struct approach we can add it as a field of the job and
// avoid hardcoded paths.
// Even better: We add the iomanager interface as a field of the job. The FileManager implements
// the iomanager interface. Then we can switch the io mechanism between the FileManager and other
// structs that implement the iomanager interface (e.g. the CMDManager).
// That allow us to relatively easy replace the input and output mechanism without
// having to edit the job code itself.

type FileManager struct {
	InputFilePath  string
	OutputFilePath string
}

func (fm FileManager) ReadLines() ([]string, error) {
	file, err := os.Open(fm.InputFilePath)

	if err != nil {
		return nil, errors.New("Failed to open file.")
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	// bufio allow us to create a scanner, which is a value that exposes methods
	// that can be used for reading content, e.g. from a file.

	// NewScanner() wants an argument of type io.Reader, which is a built-in interface
	// that is implemented by the *os.File returned by the os.Open() function.
	// i.e. *os.File satisfies the io.Reader interface.

	// scanner.Scan()
	// - Moves forward one line and reads the line that we were.
	// - Returns a bool, which is false if there is nothing left to scan.

	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = scanner.Err() // it will find out if a error occured earlier

	if err != nil {
		// file.Close()
		return nil, errors.New("Failed to read line in file.")
	}

	// file.Close()
	return lines, nil
}

func (fm FileManager) WriteResult(data interface{}) error {
	file, err := os.Create(fm.OutputFilePath)

	if err != nil {
		return errors.New("Failed to create file.")
	}

	defer file.Close()

	time.Sleep(3 * time.Second)

	encoder := json.NewEncoder(file)
	// the encoder will be used to convert values to text that follows JSON format
	// NewEncoder() accepts an interface which is implemented by the file type (*os.File)
	err = encoder.Encode(data)

	if err != nil {
		// file.Close()
		return errors.New("Faild to convert data to JSON.")
	}

	// file.Close()
	return nil
}

func New(inputPath, outputPath string) FileManager {
	return FileManager{
		InputFilePath:  inputPath,
		OutputFilePath: outputPath,
	}
}
