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

type PodcastEntry struct {
	WrapperType            string   `json:"wrapperType"`
	Kind                   string   `json:"kind"`
	ArtistId               int      `json:"artistId"`
	CollectionId           int      `json:"collectionId"`
	TrackId                int      `json:"trackId"`
	ArtistName             string   `json:"artistName"`
	CollectionName         string   `json:"collectionName"`
	TrackName              string   `json:"trackName"`
	CollectionCensoredName string   `json:"collectionCensoredName"`
	TrackCensoredName      string   `json:"trackCensoredName"`
	ArtistViewUrl          string   `json:"artistViewUrl"`
	CollectionViewUrl      string   `json:"collectionViewUrl"`
	FeedUrl                string   `json:"feedUrl"`
	TrackViewUrl           string   `json:"trackViewUrl"`
	ArtworkUrl30           string   `json:"artworkUrl30"`
	ArtworkUrl60           string   `json:"artworkUrl60"`
	ArtworkUrl100          string   `json:"artworkUrl100"`
	CollectionPrice        float64  `json:"collectionPrice"`
	TrackPrice             float64  `json:"trackPrice"`
	TrackRentalPrice       float64  `json:"trackRentalPrice"`
	CollectionHdPrice      float64  `json:"collectionHdPrice"`
	TrackHdPrice           float64  `json:"trackHdPrice"`
	TrackHdRentalPrice     float64  `json:"trackHdRentalPrice"`
	ReleaseDate            string   `json:"releaseDate"`
	CollectionExplicitness string   `json:"collectionExplicitness"`
	TrackExplicitness      string   `json:"trackExplicitness"`
	TrackCount             int      `json:"trackCount"`
	Country                string   `json:"country"`
	Currency               string   `json:"currency"`
	PrimaryGenreName       string   `json:"primaryGenreName"`
	ContentAdvisoryRating  string   `json:"contentAdvisoryRating"`
	ArtworkUrl600          string   `json:"artworkUrl600"`
	GenreIds               []string `json:"genreIds"`
	Genres                 []string `json:"genres"`
}

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
	saveTreshold := 20

	print("Creating payload array...")
	payloads := make([]LookupResponse, 0)
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

				go concurrentLookup(batch, schedulerChannel)
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
					payloads = append(payloads, parsed)
					// println("Done")

					if len(payloads) >= saveTreshold {
						println("Saving", len(payloads), "payloads to file...")

						resultString := "Result counts: "
						for i, item := range payloads {
							resultString += fmt.Sprint(item.ResultCount)
							if i != len(payloads)-1 {
								resultString += ","
							}
						}
						println(resultString)

						println("Payloads saved")

						payloads = make([]LookupResponse, 0)
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
	var wg sync.WaitGroup
	responses := make(chan WaitGroupMessage, len(urls))

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

	println(len(urls), " concurrent requests fired. Awaiting responses...")

	wg.Wait()
	close(responses)

	println(len(urls), "responses received")
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

		channel <- SchedulerMessage{
			Success: response.Success,
			Command: schedulerCommand,
			Task: SchedulerTask{
				Url:     response.Url,
				Payload: response.Payload,
			},
		}
	}

	println("Responses processed")
	println(successes, "successful and", failures, "failed responses. Total:", successes+failures, "/", len(urls))
	println("Retries will be scheduled for the", failures, "failed requests")
}
