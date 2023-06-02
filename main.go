package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/scheduler"
	"github.com/bigusbeckus/podcast-feed-fetcher/utils"
)

const PODCAST_URLS string = "data/podcasts.txt"
const PODCAST_LOOKUP_URL_BASE string = "https://itunes.apple.com/lookup?entity=podcast&id="

const PODCAST_SINGLE_URL_IDS_COUNT int = 100
const PODCAST_CONCURRENT_LOOKUPS_COUNT int = 100

func Init() {
	println("Running migrations...")
	models.RunMigrations()
	println("Migrations done")
}

func main() {
	fmt.Println("Feed fetcher started")

	Init()

	// Read podcast list from file
	content, err := os.ReadFile(PODCAST_URLS)

	if err != nil {
		log.Fatal(err)
	}

	contentLines := strings.Split(strings.TrimSpace(string(content)), "\n")
	fmt.Printf("%d podcasts read from file `%s`\n", len(contentLines), PODCAST_URLS)

	// Extract IDs
	print("Extracting IDs from URLs...")
	for i, value := range contentLines {
		contentLines[i] = utils.ExtractPodcastId(value)
	}
	println("Done")

	// Create batch lookup URLs
	batchSize := PODCAST_SINGLE_URL_IDS_COUNT
	print("Creating lookup URLs with batch size " + fmt.Sprint(batchSize) + "...")
	batchLinks := utils.CreateBatchLookupLinks(PODCAST_LOOKUP_URL_BASE, contentLines, 100)
	println("Done")

	fmt.Printf("%d batched links generated\n", len(batchLinks))
	// println(strings.Join(batchLinks, "\n\n"))

	scheduler.Start(batchLinks, PODCAST_CONCURRENT_LOOKUPS_COUNT)
}
