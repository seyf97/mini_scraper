package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/seyf97/mini_scraper/utils"
	"golang.org/x/net/html"
)

type Job struct {
	id  int
	url string
}

type Result struct {
	job   Job
	title string
	err   error
}

// Timeout for the get request
const TIMEOUT time.Duration = 10 * time.Second

// Initialized after determining the
var NUM_WORKERS int

var jobs = make(chan Job, NUM_WORKERS)
var results = make(chan Result, NUM_WORKERS)

// Returns the text from a title element, if it exists
func get_title(n *html.Node) string {
	if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
		return n.FirstChild.Data
	}
	return ""
}

// Breadth First Search
func traverseNodes(root *html.Node) string {
	queue := []*html.Node{root}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.Type == html.ElementNode && current.Data == "title" {
			title := get_title(current)
			if title != "" {
				return title
			}
		}

		for child := current.FirstChild; child != nil; child = child.NextSibling {
			queue = append(queue, child)
		}
	}

	return ""
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

	title := traverseNodes(doc)

	return title, nil
}

// Processes jobs from the job chan and sends results to result chan.
//
// Signals WorkerPool once jobs are depleted
func worker(wg *sync.WaitGroup) {
	defer wg.Done()
	for job := range jobs {

		title, err := processPage(job.url)
		res := Result{job: job,
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
		job := Job{id: i, url: link}
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
func get_domains(links []string) map[string]bool {
	domains := map[string]bool{}

	for _, link := range links {
		u, err := url.Parse(link)
		if err != nil {
			panic(err)
		}

		// Add the unique domain
		_, ok := domains[u.Host]
		if !ok {
			domains[u.Host] = true
		}

	}
	return domains
}

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

	// Determine num_workers and scrape
	domains := get_domains(links)

	NUM_WORKERS = len(domains)

	start := time.Now()

	done_chan := make(chan bool)

	go allocate_jobs(links)
	go collect_results(done_chan)
	createWorkerPool(NUM_WORKERS)

	<-done_chan

	end := time.Now()

	diff_seconds := end.Sub(start).Seconds()

	fmt.Printf("visited %v links in %v seconds using %v workers\n", len(links), diff_seconds, len(domains))
}
