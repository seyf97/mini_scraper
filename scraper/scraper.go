package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type job struct {
	domain string
	urls   []string
}

type result struct {
	domain   string
	url      string
	finalURL string
	err      error
}

type Scraper struct {
	jobs    chan job
	results chan result
}

// Constants
const TIMEOUT time.Duration = 10 * time.Second
const DELAY time.Duration = 500 * time.Millisecond
const MAX_WORKERS int = 50000

// Initialized after determining the
var NUM_WORKERS int

func NewScraper(numWorkers int) *Scraper {
	return &Scraper{
		jobs:    make(chan job, numWorkers),
		results: make(chan result, numWorkers),
	}
}

// Gets the page title
func processPage(link string) (string, error) {
	client := http.Client{Timeout: TIMEOUT}
	resp, err := client.Get(link)
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

			res := result{
				domain:   job.domain,
				finalURL: finalURL,
				url:      url,
				err:      err,
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
func (s *Scraper) collect_results(done_chan chan bool) {
	for res := range s.results {
		fmt.Printf("url_i: %v\nurl_f: %v\nerror: %v\n\n", res.url, res.finalURL, res.err)
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
func Run(links []string) {

	// Get links per domain
	domainLinks := getDomainLinks(links)

	if len(domainLinks) > MAX_WORKERS {
		NUM_WORKERS = MAX_WORKERS
	} else {
		NUM_WORKERS = len(domainLinks)
	}

	// Init chans
	s := *NewScraper(NUM_WORKERS)

	done_chan := make(chan bool)

	go s.allocate_jobs(domainLinks)
	go s.collect_results(done_chan)
	s.createWorkerPool(NUM_WORKERS)

	<-done_chan
}
