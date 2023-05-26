package scheduler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"sync"
	"time"
)

// Queue - Dequeue

// Rate limit

// Parallel fetch

// Retry
// func Start(urls []string, fetchBatchSize int) {
//
//	func Start() {
//		// queue := urls
//		ticker := time.NewTicker(3 * time.Second)
//		defer ticker.Stop()
//		done := make(chan bool)
//		go func() {
//			println("the go func? idk")
//	    time.Sleep(30 * time.Second)
//			done <- true
//		}()
//		for {
//			select {
//			case <-done:
//				println("Done!")
//				return
//			case t := <-ticker.C:
//				fmt.Println("Current time: ", t)
//			}
//		}
//
// }
type SchedulerCommand int

const (
	Retry SchedulerCommand = iota
	Stop
	Save
)

type SchedulerTask struct {
	Url     string
	Payload map[string]interface{}
}

type SchedulerMessage struct {
	Success bool
	Task    SchedulerTask
	Command SchedulerCommand
}

func Start(urls []string, fetchBatchSize int) {
	fmt.Printf("Scheduler started with batch size %d", fetchBatchSize)
	print("\n")

	print("Starting ticker...")
	ticker := time.NewTicker(3 * time.Second) // 3 seconds is the iTunes API rate limit
	println("Done")

	print("Creating SchedulerMessage channel...")
	schedulerChannel := make(chan SchedulerMessage)
	println("Done")

	lookupUrls := urls
	saveTreshold := 20

	print("Creating payload array...")
	payloads := make([]map[string]interface{}, 0)
	println("Done")

	totalProcessed := 0
	for {
		select {
		case t := <-ticker.C:
			fmt.Println("Tick at: ", t)
			if len(lookupUrls) > 0 {
				fmt.Printf("Sending %d requests...", fetchBatchSize)
				batch := extractBatch(&lookupUrls, fetchBatchSize)
				// concurrentLookup(batch, schedulerChannel)
				concurrentLookup(batch)
				println("Done")
			}
		case msg := <-schedulerChannel:
			if totalProcessed >= len(urls) || msg.Command == Stop {
				// Stop processing (Stop command or finished processing)
				print("Finished: ")

				if msg.Command == Stop {
					println("Received 'Stop' message")
				} else {
					println("All URLs processed")
				}

				print("Stopping ticker...")
				ticker.Stop()
				println("Done")
			} else if msg.Command == Retry {
				// Retry on failure
				print("Lookup failed. Adding URL back to queue...")
				lookupUrls = append(lookupUrls, msg.Task.Url)
				println("Done")
			} else if msg.Success {
				// Success
				print("URL successfully processed. Adding payload to list...")
				totalProcessed++
				payloads = append(payloads, msg.Task.Payload)
				println("Done")

				if len(payloads) >= saveTreshold {
					println("Saving payloads to file...")

					print("Result counts: ")
					for _, item := range payloads {
						print(fmt.Sprint(item["resultCount"]), ",")
					}
					println("Done")

					payloads = make([]map[string]interface{}, 0)
				}
			} else {
				println("Unhandled message")
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

// func concurrentLookup(urls []string, channel chan SchedulerMessage) {
func concurrentLookup(urls []string) {
	// for _, url := range urls {
	// 	go doLookup(url, channel)
	// }
	var wg sync.WaitGroup
	responses := make(chan string, len(urls))

	for i, url := range urls {
		wg.Add(1)
		go func(url string, index int) {
			defer wg.Done()

			print("Fetching data...")
			resp, err := http.Get(url)
			if err != nil || resp.Status != "200 OK" {
				// Requeue on fetch error
				responses <- fmt.Sprintf("Error fetching url %s. Retry scheduled.\n", url)
				return
			}
			defer resp.Body.Close()

			fmt.Printf("%d - Done. Status: %s\n", index, resp.Status)
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				responses <- fmt.Sprintf("Error reading response body from %s: %s", url, err.Error())
				return
			}

			responses <- fmt.Sprintf("Response from %s:\n%s", url, string(body))
		}(url, i+1)
	}

	wg.Wait()
	close(responses)

	for response := range responses {
		fmt.Println(response)
	}
}

func doLookup(url string, channel chan SchedulerMessage) {
	print("Fetching data...")
	resp, err := http.Get(url)
	if err != nil || resp.Status != "200 OK" {
		// Requeue on fetch error
		channel <- SchedulerMessage{
			Success: false,
			Command: Retry,
			Task: SchedulerTask{
				Url: url,
			},
		}
		fmt.Printf("Error fetching url %s. Retry scheduled.\n", url)
		return
	}
	defer resp.Body.Close()

	println("Done. Status:", resp.Status)

	// respStream, err := ioutil.ReadAll(resp.Body)
	// if err == nil {
	// 	log.Errorf("Body parse with io.ReadAll error: %s", err)
	// } else {
	// 	println("Response:", string(respStream))
	// }

	print("Parsing JSON response...")
	var contents map[string]interface{}
	jsonErr := json.NewDecoder(resp.Body).Decode(&contents)
	if jsonErr != nil {
		println("Error")
		println("Parsing returned JSON failed")
		return
	}
	println("Done")

	// if contents["resultCount"] == nil {
	// 	fmt.Println("Nil results found:")
	// 	respStream, err := ioutil.ReadAll(resp.Body)
	// 	if err == nil {
	// 		log.Errorf("Body parse with io.ReadAll error: %s", err)
	// 	} else {
	// 		println("Response:", string(respStream))
	// 	}
	// }

	channel <- SchedulerMessage{
		Success: true,
		Command: Save,
		Task: SchedulerTask{
			Url:     url,
			Payload: contents,
		},
	}

	/* body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// No retry on body parsing error
		channel <- SchedulerMessage{
			Task: SchedulerTask{
				Url:     url,
				Success: false,
				Retry:   false,
			},
		}
		log.Errorf("Error parsing response body for url %s. No retry will be attempted.", url)
		return
	} */
}
