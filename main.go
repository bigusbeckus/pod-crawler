package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/utils"
)

const PODCAST_URLS string = "data/podcasts.txt"

func main() {
	fmt.Println("Feed fetcher started")

	content, err := os.ReadFile(PODCAST_URLS)

	if err != nil {
		log.Fatal(err)
	}

	contentLines := strings.Split(strings.TrimSpace(string(content)), "\n")
	fmt.Printf("%d podcasts read from file `%s`\n", len(contentLines), PODCAST_URLS)

	var input string
	fmt.Println("Would you like to list all podcasts?")
	fmt.Println("'Y' or 'yes' to list podcasts, any other input to cancel")
	fmt.Scan(&input)
	normalizedInput := strings.ToLower(input)
	if utils.ArrayIncludes([]string{"yes", "y"}, normalizedInput) {
		for i, value := range contentLines {
			println(i, value)
		}
	} else {
		fmt.Println("Exiting...")
		os.Exit(0)
	}
}
