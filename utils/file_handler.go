package utils

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/seyf97/mini_scraper/scraper"
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

func WriteCSV(filename string, results []scraper.Result) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	err = writer.Write([]string{"link", "redirected_link", "error"})
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write the rest
	for _, res := range results {
		var err_val string
		if res.Err != nil {
			err_val = fmt.Sprintf("%v", res.Err)
		} else {
			err_val = ""
		}

		err := writer.Write([]string{res.Url, res.FinalURL, err_val})
		if err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}

	}
	return nil
}

// Gets the file name using flags
func GetFileNames() (string, string) {
	var fileNameI, fileNameO string
	flag.StringVar(&fileNameI, "i", "", "Path to the input CSV file")
	flag.StringVar(&fileNameO, "o", "", "Path to the output CSV file")

	flag.Parse()

	if fileNameI == "" {
		panic("Input flag has to be provided using: -i in_file.csv")
	}

	if fileNameO == "" {
		panic("Ouput flag has to be provided using: -o out_file.csv")
	}

	return fileNameI, fileNameO
}
