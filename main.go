package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/utils"
)

const PODCAST_URLS string = "data/podcasts.txt"

func extractPodcastId(podcastUrl string) string {
	if !utils.StringIncludes(podcastUrl, "/") {
		panic("Not a valid url")
	}
	parts := strings.Split(podcastUrl, "/")
	id := parts[len(parts)-1]
	if id[:2] == "id" {
		return id[2:]
	} else {
		return id
	}
}

func main() {
	fmt.Println("Feed fetcher started")

	// Read podcast list from file
	content, err := os.ReadFile(PODCAST_URLS)

	if err != nil {
		log.Fatal(err)
	}

	contentLines := strings.Split(strings.TrimSpace(string(content)), "\n")
	fmt.Printf("%d podcasts read from file `%s`\n", len(contentLines), PODCAST_URLS)

	var input string
	var normalizedInput string

	fmt.Println("Would you like to list all podcasts?")
	fmt.Println("'Y' or 'yes' to list podcasts, any other input to cancel")
	fmt.Scan(&input)
	normalizedInput = strings.ToLower(input)
	if !utils.ArrayIncludes([]string{"yes", "y"}, normalizedInput) {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	// Output
	println(strings.Join(contentLines, "\n"))

	input = ""
	normalizedInput = ""

	fmt.Println("Would you like to extract all podcast IDs?")
	fmt.Println("'Y' or 'yes' to extract and list podcast IDs, any other input to cancel")
	fmt.Scan(&input)
	normalizedInput = strings.ToLower(input)
	if !utils.ArrayIncludes([]string{"yes", "y"}, normalizedInput) {
		fmt.Println("Exiting...")
		os.Exit(0)
	}

	// Extract IDs
	print("Extracting IDs from URLs...")
	for i, value := range contentLines {
		contentLines[i] = extractPodcastId(value)
	}
	println("Done")

	// Output
	println(strings.Join(contentLines, "\n"))
}
