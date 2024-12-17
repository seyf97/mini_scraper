package scraper

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type job struct {
	domain string
	urls   []string
}

type Result struct {
	Domain   string
	Url      string
	FinalURL string
	Err      error
}

type Scraper struct {
	jobs    chan job
	results chan Result
}

// Constants
const TIMEOUT time.Duration = 10 * time.Second
const DELAY time.Duration = 500 * time.Millisecond

var httpClient = &http.Client{Timeout: TIMEOUT}

// const MAX_WORKERS int = 5000

var USER_AGENTS = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_6_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.5938.132 Safari/537.36",
	"Mozilla/5.0 (Linux; Android 12; SM-G991B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 16_5 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.0 Mobile/15E148 Safari/604.1",
	"Mozilla/5.0 (Windows NT 10.0; WOW64; rv:117.0) Gecko/20100101 Firefox/117.0",
	"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:117.0) Gecko/20100101 Firefox/117.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 12_5_1) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
	"Mozilla/5.0 (Linux; Android 13; Pixel 7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Mobile Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:116.0) Gecko/20100101 Firefox/116.0",
	"Mozilla/5.0 (iPad; CPU OS 15_6 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/15.6 Mobile/15E148 Safari/604.1",
}

// Initialized after determining the
var NUM_WORKERS int

func NewScraper(numWorkers int) *Scraper {
	return &Scraper{
		jobs:    make(chan job, numWorkers),
		results: make(chan Result, numWorkers),
	}
}

// Gets the page title
func processPage(link string) (string, error) {

	// Get a random user agent
	randIdx := rand.Intn(len(USER_AGENTS))
	randAgent := USER_AGENTS[randIdx]

	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", randAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// Get the final redirected link
	finalURL := resp.Request.URL.String()

	// Get the title
	// doc, err := html.Parse(resp.Body)
	// if err != nil {
	// 	return "", err
	// }

	// title := get_title(doc)

	return finalURL, nil
}

// Processes jobs from the job chan and sends results to result chan.
//
// Signals WorkerPool once jobs are depleted
func (s *Scraper) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range s.jobs {

		for _, url := range job.urls {
			finalURL, err := processPage(url)

			res := Result{
				Domain:   job.domain,
				FinalURL: finalURL,
				Url:      url,
				Err:      err,
			}

			s.results <- res
		}

		// Sleep between each request
		time.Sleep(DELAY)
	}
}

// Init a worker pool where each worker gets a job from job chan
func (s *Scraper) createWorkerPool(num_workers int) {
	var wg sync.WaitGroup

	for i := 0; i < num_workers; i++ {
		wg.Add(1)
		go s.worker(&wg)
	}
	wg.Wait()
	close(s.results)
}

// Sends jobs to job chan
func (s *Scraper) allocate_jobs(domLinks map[string][]string) {

	for domain, urls := range domLinks {
		job := job{domain: domain, urls: urls}
		s.jobs <- job
	}
	close(s.jobs)
}

// Collects results from results chan
func (s *Scraper) collect_results(done_chan chan bool, out_results *[]Result) {
	for res := range s.results {
		*out_results = append(*out_results, res)
		if res.Err != nil {
			fmt.Printf("url_i: %v\nurl_f: %v\nerror: %v\n\n", res.Url, res.FinalURL, res.Err)
		} else {
			fmt.Printf("url_i: %v\nurl_f: %v\nerror: \n\n", res.Url, res.FinalURL)
		}
	}
	done_chan <- true
}

// Gets unique hosts from a given list of links
func getDomainLinks(links []string) map[string][]string {
	domainLinks := map[string][]string{}

	for _, link := range links {
		u, err := url.Parse(link)
		if err != nil {
			panic(err)
		}

		// Add the unique domain
		_, isPresent := domainLinks[u.Host]
		if !isPresent {
			domainLinks[u.Host] = []string{link}
		} else {
			domainLinks[u.Host] = append(domainLinks[u.Host], link)
		}

	}
	return domainLinks
}

// Scrapes links concurrently
func Run(links []string) (results []Result) {

	// Get links per domain
	domainLinks := getDomainLinks(links)

	var out_results []Result

	// if len(domainLinks) > MAX_WORKERS {
	// 	NUM_WORKERS = MAX_WORKERS
	// } else {
	// 	NUM_WORKERS = len(domainLinks)
	// }
	NUM_WORKERS := len(domainLinks)

	// Init chans
	s := *NewScraper(NUM_WORKERS)

	done_chan := make(chan bool)

	go s.allocate_jobs(domainLinks)
	go s.collect_results(done_chan, &out_results)
	s.createWorkerPool(NUM_WORKERS)

	<-done_chan
	return out_results
}

func BatchRun(fileInName, fileOutName string, batchSize int) {

	fileIn, err := os.Open(fileInName)
	if err != nil {
		panic(fmt.Sprintf("error opening file: %v\n", err))
	}
	defer fileIn.Close()

	fileOut, err := os.OpenFile(fileOutName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(fmt.Sprintf("error opening output file: %v\n", err))
	}
	defer fileOut.Close()

	writer := csv.NewWriter(fileOut)
	defer writer.Flush()

	// Header
	err = writer.Write([]string{"link", "redirected_link", "error"})
	if err != nil {
		panic(fmt.Sprintf("failed to write header: %v\n", err))
	}

	reader := csv.NewReader(fileIn)

	urls := []string{}
	for {
		row, err := reader.Read()
		if err != nil {
			// Reached the end of the line
			if err.Error() == "EOF" {
				break
			}
			panic(fmt.Sprintf("error reading file: %v\n", err))
		}

		urls = append(urls, row[0])

		// Process after batch size is reached
		if len(urls) >= batchSize {
			processBatch(urls, writer)
			urls = []string{} // clear urls
		}
	}

	if len(urls) > 0 {
		processBatch(urls, writer)
	}

	fmt.Println("Finished processing")

}

func processBatch(urls []string, writer *csv.Writer) {
	fmt.Println("Starting batch processing...")
	time.Sleep(1 * time.Second)

	results := Run(urls)

	for _, res := range results {
		var err_val string
		if res.Err != nil {
			err_val = fmt.Sprintf("%v", res.Err)
		} else {
			err_val = ""
		}

		err := writer.Write([]string{res.Url, res.FinalURL, err_val})
		if err != nil {
			panic(fmt.Sprintf("failed to write record: %v\n", err))
		}

	}
	writer.Flush()
	err := writer.Error()

	if err != nil {
		panic(fmt.Sprintf("error flushing writer: %v\n", err))

	}

	fmt.Println("Batch processing complete.")

}
