package main

import (
	"fmt"
	"time"

	// _ "net/http/pprof"

	"github.com/seyf97/mini_scraper/scraper"
	"github.com/seyf97/mini_scraper/utils"
)

func main() {

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	// Read file and get urls
	fileNameIn, fileNameOut := utils.GetFileNames()

	fmt.Printf("Reading file: %s\n", fileNameIn)

	// Scrape links
	start := time.Now()
	batchSize := 5000
	scraper.BatchRun(fileNameIn, fileNameOut, batchSize)
	end := time.Now()

	diff_seconds := end.Sub(start).Seconds()

	fmt.Printf("Finished in %v seconds\n", diff_seconds)
}
