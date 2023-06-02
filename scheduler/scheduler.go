package scheduler

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/bigusbeckus/podcast-feed-fetcher/models"
)

type PodcastEntry = models.PodcastEntry
type LookupResponse struct {
	ResultCount int            `json:"resultCount"`
	Results     []PodcastEntry `json:"results"`
}

// Queue - Dequeue

// Rate limit

// Parallel fetch

// Retry
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

func Start(urls []string, fetchBatchSize int) {
	fmt.Printf("Scheduler started with batch size %d\n", fetchBatchSize)

	print("Starting ticker...")
	ticker := time.NewTicker(3 * time.Second) // 3 seconds is the iTunes API rate limit
	println("Done")

	print("Creating SchedulerMessage channel...")
	schedulerChannel := make(chan SchedulerMessage)
	println("Done")

	lookupUrls := urls
	saveTreshold := 200

	print("Creating payload array...")
	payloads := make([]PodcastEntry, 0)
	// resultCounts := make([]int, 0)
	println("Done")

	totalProcessed := 0
	for {
		select {
		case t := <-ticker.C:
			fmt.Println("Tick at: ", t)
			if len(lookupUrls) > 0 {
				fmt.Printf("Sending %d requests...\n", fetchBatchSize)

				print("Extracting batch...")
				batch := extractBatch(&lookupUrls, fetchBatchSize)
				println("Done")

				concurrentLookup(batch, schedulerChannel)
			}
		case msg := <-schedulerChannel:
			if msg.Command == Stop {
				// Stop processing (Stop command or finished processing)
				if totalProcessed >= len(urls) {
					println("Complete! All URLs processed")
				} else {
					println("Received premature 'Stop' message")
				}

				println("Stopping ticker...")
				ticker.Stop()
				println("Ticker stopped")
			} else if msg.Command == Retry {
				// Retry on failure
				// print("Lookup failed. Adding URL back to queue...")
				lookupUrls = append(lookupUrls, msg.Task.Url)
				// println("Done")
			} else if msg.Success {
				// Success
				// print("URL successfully processed. Adding payload to list...")
				go func() {
					var parsed LookupResponse
					jsonErr := json.Unmarshal([]byte(string(msg.Task.Payload)), &parsed)
					if jsonErr != nil {
						println(fmt.Sprintf("Error parsing response body from %s: %s", msg.Task.Url, jsonErr.Error()))
						lookupUrls = append(lookupUrls, msg.Task.Url)
						return
					}
					totalProcessed++

					// resultCounts = append(resultCounts, parsed.ResultCount)
					for _, result := range parsed.Results {
						payloads = append(payloads, result)
					}
					// println("Done")

					if len(payloads) >= saveTreshold {
						println("Saving", len(payloads), "payloads to file...")

						// resultString := "Result counts: "
						// for i, resultCount := range resultCounts {
						// 	resultString += fmt.Sprint(resultCount)
						// 	if i != len(resultCounts)-1 {
						// 		resultString += ","
						// 	}
						// }
						// println(resultString)

						payloadsJson, err := json.Marshal(payloads)
						if err != nil {
							panic("Payload JSON conversion failed")
						}

						os.WriteFile("output/payloads__"+time.Now().Format(time.RFC3339)+".json", []byte(payloadsJson), fs.ModeDevice)

						println("Payloads saved")

						payloads = make([]PodcastEntry, 0)
						// resultCounts = make([]int, 0)
					}
				}()
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

func concurrentLookup(urls []string, channel chan SchedulerMessage) {
	urlsCount := len(urls)
	var wg sync.WaitGroup
	responses := make(chan WaitGroupMessage, urlsCount)

	for _, url := range urls {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			// print("Fetching data...")
			resp, err := http.Get(url)
			if err != nil || resp.Status != "200 OK" {
				// Requeue on fetch error
				responses <- WaitGroupMessage{
					Success: false,
					Url:     url,
				}
				// println(fmt.Sprintf("Error fetching url %s. Retry scheduled.\n", url))
				return
			}
			defer resp.Body.Close()

			// fmt.Printf("Done. Status: %s", resp.Status)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				responses <- WaitGroupMessage{
					Success:      false,
					HttpResponse: resp.Status,
					Url:          url,
				}
				// println(fmt.Sprintf("Error reading response body from %s: %s", url, err.Error()))
				return
			}

			responses <- WaitGroupMessage{
				Success:      true,
				HttpResponse: resp.Status,
				Url:          url,
				Payload:      string(body),
			}
			// println(fmt.Sprintf("Response from %s:\n%s", url, string(body)))
		}(url)
	}

	println(urlsCount, " concurrent requests fired. Awaiting responses...")

	wg.Wait()
	close(responses)

	println(urlsCount, "responses received")

	go func(responses chan WaitGroupMessage, schedulerChannel chan SchedulerMessage, urlsCount int) {
		println("Response processing started...")

		successes := 0
		failures := 0

		for response := range responses {
			var schedulerCommand SchedulerCommand
			if response.Success {
				schedulerCommand = Save
				successes++
			} else {
				schedulerCommand = Retry
				failures++
			}

			schedulerChannel <- SchedulerMessage{
				Success: response.Success,
				Command: schedulerCommand,
				Task: SchedulerTask{
					Url:     response.Url,
					Payload: response.Payload,
				},
			}
		}

		println("Responses processed")
		println(successes, "successful and", failures, "failed responses. Total:", successes+failures, "/", urlsCount)
		println("Retries will be scheduled for the", failures, "failed requests")
	}(responses, channel, urlsCount)
}
