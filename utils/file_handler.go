package utils

import (
	"encoding/csv"
	"os"
)

// Reads the csv with the links, assuming its a single column
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
