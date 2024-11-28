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

// Timeout for the get request
const TIMEOUT time.Duration = 10 * time.Second

// Initialized after determining the
var NUM_WORKERS int

var jobs = make(chan job, NUM_WORKERS)
var results = make(chan result, NUM_WORKERS)

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
func worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {

		title, err := processPage(job.url)
		res := result{job: job,
			title: title,
			err:   err}

		results <- res

		// Sleep between each request
		time.Sleep(1 * time.Second)
	}
}

// Init a worker pool where each worker gets a job from job chan
func createWorkerPool(num_workers int) {
	var wg sync.WaitGroup

	for i := 0; i < num_workers; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(results)
}

// Sends jobs to job chan
func allocate_jobs(links []string) {

	for i, link := range links {
		job := job{id: i, url: link}
		jobs <- job
	}
	close(jobs)
}

// Collects results from results chan
func collect_results(done_chan chan bool) {
	for res := range results {
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
func Scrape(links []string) {

	// Determine num_workers
	domainLinks := getDomainLinks(links)
	NUM_WORKERS = len(domainLinks)

	done_chan := make(chan bool)

	go allocate_jobs(links)
	go collect_results(done_chan)
	createWorkerPool(NUM_WORKERS)

	<-done_chan
}
