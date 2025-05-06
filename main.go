package main

import (
	"fmt"
	"os"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Printf("no website provided\n")
		os.Exit(1)
	} else if len(args) > 1 {
		fmt.Printf("too many arguments provided\n")
		os.Exit(1)
	}

	fmt.Printf("starting crawl of: %s\n", args[0])

	linksVisited := make(map[string]int)

	crawlPage(args[0], args[0], linksVisited)

	fmt.Printf(" Visits -------    Links\n\n")
	for link, numVisits := range linksVisited {
		fmt.Printf("   %d\t\t%s\n", numVisits, link)
	}
}
