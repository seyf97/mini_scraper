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
	fileName := utils.GetFileName()

	fmt.Printf("Reading file: %s\n", fileName)

	links, err := utils.ReadCSV(fileName, false)
	if err != nil {
		panic(err)
	}

	if len(links) == 0 {
		panic(errors.New("input file has no links"))
	}

	// Scrape links
	start := time.Now()
	scraper.Scrape(links)
	end := time.Now()

	diff_seconds := end.Sub(start).Seconds()

	fmt.Printf("Visited %v links in %v seconds\n", len(links), diff_seconds)
}
