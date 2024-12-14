package main

import (
	"errors"
	"fmt"
	"time"

	"github.com/seyf97/mini_scraper/scraper"
	"github.com/seyf97/mini_scraper/utils"
)

func main() {
	// Read file and get urls
	fileNameIn, fileNameOut := utils.GetFileNames()

	fmt.Printf("Reading file: %s\n", fileNameIn)

	links, err := utils.ReadCSV(fileNameIn, true)
	if err != nil {
		panic(err)
	}

	if len(links) == 0 {
		panic(errors.New("input file has no links"))
	}

	// // testing
	// links = links[:10]

	// Scrape links
	start := time.Now()
	results := scraper.Run(links)
	end := time.Now()

	diff_seconds := end.Sub(start).Seconds()

	fmt.Printf("Visited %v links in %v seconds\n", len(links), diff_seconds)

	fmt.Printf("Writing file...\n")

	err = utils.WriteCSV(fileNameOut, results)
	if err != nil {
		panic(err)
	}
}
