package main

import (
	"fmt"
	"log"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/seyf97/mini_scraper/scraper"
	"github.com/seyf97/mini_scraper/utils"
)

func main() {

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// Read file and get urls
	fileNameIn, fileNameOut := utils.GetFileNames()

	fmt.Printf("Reading file: %s\n", fileNameIn)

	// links, err := utils.ReadCSV(fileNameIn, true)
	// if err != nil {
	// 	panic(err)
	// }

	// if len(links) == 0 {
	// 	panic(errors.New("input file has no links"))
	// }

	// Scrape links
	start := time.Now()
	// results := scraper.Run(links)
	batchSize := 100
	scraper.BatchRun(fileNameIn, fileNameOut, batchSize)
	end := time.Now()

	diff_seconds := end.Sub(start).Seconds()

	// fmt.Printf("Visited %v links in %v seconds\n", len(links), diff_seconds)
	fmt.Printf("Finished in %v seconds\n", diff_seconds)

	// fmt.Printf("Writing file...\n")

	// err = utils.WriteCSV(fileNameOut, results)
	// if err != nil {
	// 	panic(err)
	// }
}
