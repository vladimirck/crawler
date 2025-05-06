package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	//"time"
)

type config struct {
	pages              map[string]int
	baseURL            string
	mu                 *sync.Mutex
	concurrencyControl chan struct{}
	wg                 *sync.WaitGroup
	maxConcurrency     int
	maxPages           int
	isMaxPageReached   bool
}

func main() {
	args := os.Args[1:]

	if len(args) < 3 {
		fmt.Printf("too few arguments\n")
		os.Exit(1)
	} else if len(args) > 3 {
		fmt.Printf("too many arguments provided\n")
		os.Exit(1)
	}

	maxThreadCount, err := strconv.Atoi(args[1])
	if err != nil {
		fmt.Printf("error reading maxThread: %c", err)
		os.Exit(1)
	}
	maxPageCount, err := strconv.Atoi(args[2])
	if err != nil {
		fmt.Printf("error reading maxPages: %c", err)
		os.Exit(1)
	}

	cfg := config{
		pages:              make(map[string]int),
		baseURL:            args[0],
		mu:                 &sync.Mutex{},
		concurrencyControl: make(chan struct{}, maxThreadCount),
		wg:                 &sync.WaitGroup{},
		maxPages:           maxPageCount,
		maxConcurrency:     maxThreadCount,
	}

	cfg.wg.Add(1)
	go cfg.crawlPage(cfg.baseURL)
	fmt.Printf("starting crawl of: %s\n\n", cfg.baseURL)

	//time.Sleep(8 * time.Second)

	cfg.wg.Wait()

	cfg.printReport()
}
