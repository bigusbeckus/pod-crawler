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

const SAVE_TRESHOLD = 50000

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
	// resultCounts := make([]int, 0)
	println("Done")

	totalProcessed := 0
	payloadsNextIndex := 0
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
			if msg.Command == Skip {
				totalProcessed++
			} else if msg.Command == Stop {
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
				lookupUrls = append(lookupUrls, msg.Task.Url)
			} else if msg.Success {
				// Success
				var parsed LookupResponse
				jsonErr := json.Unmarshal([]byte(string(msg.Task.Payload)), &parsed)
				if jsonErr != nil {
					println(fmt.Sprintf("Error parsing response body from %s: %s", msg.Task.Url, jsonErr.Error()))
					lookupUrls = append(lookupUrls, msg.Task.Url)
					return
				}
				totalProcessed++

				payloadsLastIndex := len(payloads) - 1
				for _, result := range parsed.Results {
					if payloadsNextIndex > payloadsLastIndex {
						payloads = append(payloads, result)
					} else {
						payloads[payloadsNextIndex] = result
					}
					payloadsNextIndex++
				}

				currentBatchCount := payloadsNextIndex
				if currentBatchCount >= SAVE_TRESHOLD {
					println("Saving", currentBatchCount, "entries to file...")

					payloadsJson, err := json.Marshal(payloads)
					if err != nil {
						panic("Payload JSON conversion failed")
					}

					os.WriteFile("output/payloads__"+time.Now().Format(time.RFC3339)+".json", []byte(payloadsJson), fs.ModeDevice)

					println("Payloads saved")

					payloads = make([]PodcastEntry, SAVE_TRESHOLD)
					payloadsNextIndex = 0
				}
				// go func(payloads []PodcastEntry, schedulerTask *SchedulerTask) {
				// 	task := *schedulerTask
				// 	var parsed LookupResponse
				// 	jsonErr := json.Unmarshal([]byte(string(task.Payload)), &parsed)
				// 	if jsonErr != nil {
				// 		println(fmt.Sprintf("Error parsing response body from %s: %s", task.Url, jsonErr.Error()))
				// 		lookupUrls = append(lookupUrls, task.Url)
				// 		return
				// 	}
				// 	totalProcessed++
				//
				// 	for _, result := range parsed.Results {
				// 		payloads = append(payloads, result)
				// 	}
				//
				// 	if len(payloads) >= SAVE_TRESHOLD {
				// 		println("Saving", len(payloads), "payloads to file...")
				//
				// 		payloadsJson, err := json.Marshal(payloads)
				// 		if err != nil {
				// 			panic("Payload JSON conversion failed")
				// 		}
				//
				// 		os.WriteFile("output/payloads__"+time.Now().Format(time.RFC3339)+".json", []byte(payloadsJson), fs.ModeDevice)
				//
				// 		println("Payloads saved")
				//
				// 		payloads = make([]PodcastEntry, 0)
				// 	}
				// }(payloads, &msg.Task)
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

func concurrentLookup(urls []string, schedulerChannel chan SchedulerMessage) {
	urlsCount := len(urls)
	var wg sync.WaitGroup
	failuresChannel := make(chan string, urlsCount)

	const INVALID_BODY = "invalid_body"

	for i, url := range urls {
		wg.Add(1)

		go func(url string, schedulerChannel chan SchedulerMessage, failuresChannel chan string, index int) {
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
		}(url, schedulerChannel, failuresChannel, i)
	}

	println(urlsCount, "concurrent requests fired. Awaiting responses...")

	wg.Wait()
	close(failuresChannel)

	println(urlsCount, "responses received")

	failures := len(failuresChannel)
	successes := urlsCount - failures

	println(successes, "successful and", failures, "failed responses. Total:", successes+failures, "/", urlsCount)
	failuresRatio := float64(len(failuresChannel)) / float64(urlsCount) * 100
	failurePercentage := math.Round((failuresRatio * 100)) / 100
	fmt.Println(failurePercentage, "% failure rate")

	if failures > 0 {
		go func(failuresChannel chan string) {
			failures = len(failuresChannel)
			noRetries := 0
			groupedFailures := make(map[string]int)
			for failure := range failuresChannel {
				if failure == INVALID_BODY {
					noRetries++
				}
				groupedFailures[failure] = groupedFailures[failure] + 1
			}

			var statsString string
			for key, value := range groupedFailures {
				statsString += fmt.Sprintf("%s: %d, ", key, value)
			}

			println("Stats - ", statsString[:len(statsString)-2])
			fmt.Printf("Retries will be scheduled for %d/%d of the failed requests\n", failures-noRetries, failures)
		}(failuresChannel)
	}
}
