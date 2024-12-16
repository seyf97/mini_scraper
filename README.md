# mini_scraper
A lightweight, concurrent scraper built using Golang

Implements a worker pool where each worker scrapes one domain

## Installation
1. Clone the Repository:

```
git clone https://github.com/seyf97/mini_scraper.git
```

2. Navigate to the Project Directory:

```
cd mini_scraper
```

3. Install Dependencies:

```
go mod download
```

## Usage
1. Prepare a CSV file with a single column containing the links to scrape

Example:
```
https://example.com
https://example.org
https://anotherexample.com
```

2. Run the scraper

Use the -i and -o flags to specify the input CSV file and target output CSV file, respectively.

```
go run main.go -i urls.csv -o results.csv
```
