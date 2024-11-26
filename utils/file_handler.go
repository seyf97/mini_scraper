package utils

import (
	"encoding/csv"
	"errors"
	"flag"
	"os"
)

// Reads a csv file, assuming it has a single column
func ReadCSV(fileName string, hasHeaders bool) ([]string, error) {

	file, err := os.Open(fileName)
	if err != nil {
		return []string{}, err
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return []string{}, err
	}

	urls := []string{}

	for _, row := range records {
		if len(row) > 0 {
			urls = append(urls, row[0])
		}
	}

	// Skip the headers
	if hasHeaders && len(urls) > 0 {
		urls = urls[1:]
	}

	return urls, nil
}

// Gets the file name using flags
func GetFileName() string {
	fileName := flag.String("file", "", "Path to the input CSV file")
	f := flag.String("f", "", "Short alias for -file")

	flag.Parse()

	if *fileName == "" && *f == "" {
		panic(errors.New("file name must be provided using -f or --file"))
	}

	if *f != "" {
		*fileName = *f
	}

	return *fileName
}
