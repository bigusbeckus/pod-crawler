package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/models"
)

type PodcastEntry = models.PodcastEntry
type LookupResponse struct {
	ResultCount int            `json:"resultCount"`
	Results     []PodcastEntry `json:"results"`
}

type SchedulerCommand int

type WaitGroupMessage struct {
	Success      bool
	HttpResponse string
	Url          string
	Payload      string
}

const (
	Retry SchedulerCommand = iota
	Stop
	Save
	Skip
)

type SchedulerTask struct {
	Url     string
	Payload string // map[string]interface{}
}

type SchedulerMessage struct {
	Success bool
	Task    SchedulerTask
	Command SchedulerCommand
}

const INVALID_BODY = "invalid_body"
const SAVE_TRESHOLD = 50000

func onTick(t time.Time, lookupUrls []string, schedulerChannel chan SchedulerMessage, fetchBatchSize int) {
	fmt.Println("Tick at: ", t)
	if len(lookupUrls) > 0 {
		fmt.Printf("Sending %d requests...\n", fetchBatchSize)

		print("Extracting batch...")
		batch := extractBatch(&lookupUrls, fetchBatchSize)
		println("Done")

		concurrentLookup(batch, schedulerChannel)
	}
}

// Retry on failure
func onSchedulerRetry(lookupUrls *[]string, msg SchedulerMessage) {
	*lookupUrls = append(*lookupUrls, msg.Task.Url)
}

// Stop processing (Stop command or finished processing)
func onSchedulerStop(ticker *time.Ticker, totalProcessed int, urlsCount int) {
	allProcessed := totalProcessed >= urlsCount

	if allProcessed {
		println("Complete! All URLs processed")
	} else {
		println("Received premature 'Stop' message")
	}

	println("Stopping ticker...")
	ticker.Stop()
	println("Ticker stopped")

}

func saveBatch(batch []PodcastEntry) error {
	println("Saving", len(batch), "entries to file...")

	batchJson, err := json.Marshal(batch)
	if err != nil {
		return errors.New("Payload JSON conversion failed")
	}

	os.WriteFile("output/payloads__"+time.Now().Format(time.RFC3339)+".json", []byte(batchJson), fs.ModeDevice)
	println("Payloads saved")

	return nil
}

func resetPayloads(payloads *[]PodcastEntry, nextIndex *int) {
	*payloads = make([]PodcastEntry, SAVE_TRESHOLD)
	*nextIndex = 0
}

func parseLookupResponse(msg SchedulerMessage) (*LookupResponse, error) {
	payload := msg.Task.Payload
	url := msg.Task.Url

	var parsed LookupResponse
	jsonErr := json.Unmarshal([]byte(payload), &parsed)
	if jsonErr != nil {
		err := errors.New(fmt.Sprintf("Error parsing response body from %s: %s", url, jsonErr.Error()))
		return nil, err
	}

	return &parsed, nil
}

// func dynamicAppendPayloads(payloads []PodcastEntry, parsed []PodcastEntry) {
// 	lastIndex := len(payloads) - 1
// }

// Success
func onSchedulerSuccess(lookupUrls *[]string, payloads []PodcastEntry, msg SchedulerMessage, payloadsNextIndex *int, totalProcessed *int) {
	parsed, err := parseLookupResponse(msg)
	if err != nil {
		println(err.Error())
		*lookupUrls = append(*lookupUrls, msg.Task.Url) // Schedule retry
		return
	}

	*totalProcessed++

	payloadsLastIndex := len(payloads) - 1
	for _, result := range parsed.Results {
		if *payloadsNextIndex > payloadsLastIndex {
			payloads = append(payloads, result)
		} else {
			payloads[*payloadsNextIndex] = result
		}
		*payloadsNextIndex++
	}

	currentBatch := *payloadsNextIndex
	if currentBatch >= SAVE_TRESHOLD {
		if err := saveBatch(payloads); err != nil {
			panic(err.Error())
		}
		resetPayloads(&payloads, payloadsNextIndex)
	}

}

func Start(urls []string, fetchBatchSize int) {
	fmt.Printf("Scheduler started with batch size %d\n", fetchBatchSize)

	print("Starting ticker...")
	ticker := time.NewTicker(3 * time.Second) // 3 seconds is the iTunes API rate limit
	println("Done")

	print("Creating SchedulerMessage channel...")
	schedulerChannel := make(chan SchedulerMessage)
	println("Done")

	lookupUrls := urls

	print("Creating payload array...")
	payloads := make([]PodcastEntry, SAVE_TRESHOLD)
	println("Done")

	totalProcessed := 0
	payloadsNextIndex := 0

	for {
		select {
		case t := <-ticker.C:
			onTick(t, lookupUrls, schedulerChannel, fetchBatchSize)
		case msg := <-schedulerChannel:
			// Setup context before moving this into a separate function
			if msg.Command == Skip {
				totalProcessed++
			} else if msg.Command == Stop {
				onSchedulerStop(ticker, totalProcessed, len(urls))
			} else if msg.Command == Retry {
				onSchedulerRetry(&lookupUrls, msg)
			} else if msg.Success {
				onSchedulerSuccess(
					&lookupUrls,
					payloads,
					msg,
					&totalProcessed,
					&payloadsNextIndex,
				)
			} else {
				println("Unhandled message")
				ticker.Stop()
			}
		}

	}
}

func extractBatch(urls *[]string, batchSize int) []string {
	endIndex := int(math.Min(float64(batchSize), float64(len(*urls))))
	currentBatch := (*urls)[:endIndex]
	*urls = (*urls)[endIndex:]
	return currentBatch
}

func fetch(wg *sync.WaitGroup, schedulerChannel chan SchedulerMessage, failuresChannel chan string, url string, index int) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil || resp.Status != "200 OK" {
		// Requeue on fetch error
		failuresChannel <- resp.Status
		go func() {
			schedulerChannel <- SchedulerMessage{
				Success: false,
				Command: Retry,
				Task: SchedulerTask{
					Url: url,
				},
			}
		}()
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// No requeue on invalid body
		failuresChannel <- INVALID_BODY
		go func() {
			schedulerChannel <- SchedulerMessage{
				Success: false,
				Command: Skip,
			}
		}()
		return
	}

	go func(schedulerChannel chan SchedulerMessage, url string, payload string) {
		schedulerChannel <- SchedulerMessage{
			Success: true,
			Command: Save,
			Task: SchedulerTask{
				Url:     url,
				Payload: payload,
			},
		}
	}(schedulerChannel, url, string(body))
}

func printSuccessRatios(successes int, failures int, urlsCount int) {
	println(successes, "successful and", failures, "failed responses. Total:", successes+failures, "/", urlsCount)

	failuresRatio := float64(failures) / float64(urlsCount) * 100
	failurePercentage := math.Round((failuresRatio * 100)) / 100

	fmt.Println(failurePercentage, "% failure rate")
}

func printFailureGroupStats(failuresCount int, failuresChannel chan string) {
	if failuresCount <= 0 {
		return
	}

	noRetriesCount := 0
	groupedFailures := make(map[string]int)
	for failureMessage := range failuresChannel {
		if failureMessage == INVALID_BODY {
			noRetriesCount++
		}
		groupedFailures[failureMessage] = groupedFailures[failureMessage] + 1
	}

	var statsString string
	for key, value := range groupedFailures {
		statsString += fmt.Sprintf("%s: %d, ", key, value)
	}

	println("Stats - ", statsString[:len(statsString)-2])
	fmt.Printf("Retries will be scheduled for %d/%d of the failed requests\n", failuresCount-noRetriesCount, failuresCount)
}

func printStats(failuresChannel chan string, urlsCount int) {
	failures := len(failuresChannel)
	successes := urlsCount - failures

	printSuccessRatios(successes, failures, urlsCount)
	go printFailureGroupStats(failures, failuresChannel)
}

func concurrentLookup(urls []string, schedulerChannel chan SchedulerMessage) {
	urlsCount := len(urls)

	var wg sync.WaitGroup
	failuresChannel := make(chan string, urlsCount)

	for i, url := range urls {
		wg.Add(1)
		go fetch(&wg, schedulerChannel, failuresChannel, url, i)
	}

	println(urlsCount, "concurrent requests fired. Awaiting responses...")

	wg.Wait()
	close(failuresChannel)

	println(urlsCount, "responses received")

	printStats(failuresChannel, urlsCount)
}
