package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type job struct {
	id  int
	url string
}

type result struct {
	job   job
	title string
	err   error
}

type Scraper struct {
	jobs    chan job
	results chan result
}

// Timeout for the get request
const TIMEOUT time.Duration = 10 * time.Second

// Max workers at any given point
const MAX_WORKERS int = 5000

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
	res, err := client.Get(link)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(res.Body)
	if err != nil {
		return "", err
	}

	title := get_title(doc)

	return title, nil
}

// Processes jobs from the job chan and sends results to result chan.
//
// Signals WorkerPool once jobs are depleted
func (s *Scraper) worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range s.jobs {

		title, err := processPage(job.url)
		res := result{job: job,
			title: title,
			err:   err}

		s.results <- res

		// Sleep between each request
		time.Sleep(1 * time.Second)
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
func (s *Scraper) allocate_jobs(links []string) {

	for i, link := range links {
		job := job{id: i, url: link}
		s.jobs <- job
	}
	close(s.jobs)
}

// Collects results from results chan
func (s *Scraper) collect_results(done_chan chan bool) {
	for res := range s.results {
		fmt.Printf("url: %v\ntitle: %v\nerror: %v\n\n", res.job.url, res.title, res.err)
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

	var mixedLinks []string

	for {
		fmt.Println("Collecting links per domain...")
		mixedLinks = []string{}

		// 1. For each domain, get the first link

		for domain, links := range domainLinks {
			if len(links) > 0 {
				mixedLinks = append(mixedLinks, links[0])
				domainLinks[domain] = links[1:]
			}
		}

		// End when no more new links
		if len(mixedLinks) == 0 {
			break
		}

		// 2. Scrape mixedLinks using workerpool

		if len(mixedLinks) > MAX_WORKERS {
			NUM_WORKERS = MAX_WORKERS
		} else {
			NUM_WORKERS = len(mixedLinks)
		}

		s := *NewScraper(NUM_WORKERS)

		done_chan := make(chan bool)

		go s.allocate_jobs(mixedLinks)
		go s.collect_results(done_chan)
		s.createWorkerPool(NUM_WORKERS)

		<-done_chan

	}

}
