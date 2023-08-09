package podcast

import (
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/structures"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
)

type FetcherCommand int

const (
	Stop FetcherCommand = iota
	Pause
	Resume
)

type FetchResponseData struct {
	Url     string
	Payload string
}

type FetchResponse struct {
	Success     bool
	IsBodyValid bool
	Status      int
	Data        FetchResponseData
}

type Fetcher struct {
	idPool            structures.Pool[uint64]
	concurrentFetches int
	maxIdsPerFetch    int

	ticker          *time.Ticker
	lastFetchEnd    time.Time
	CommandChannel  chan FetcherCommand
	ResponseChannel chan FetchResponse
	fetchWaitGroup  sync.WaitGroup

	pause bool
}

func (f *Fetcher) Append(ids ...uint64) {
	f.idPool.Put(ids...)
}

func (f *Fetcher) Shuffle() {
	f.idPool.Shuffle()
}

func NewFetchResponse(
	success bool,
	status int,
	isBodyValid bool,
	url string,
	payload string,
) FetchResponse {
	return FetchResponse{
		Success:     success,
		IsBodyValid: isBodyValid,
		Status:      status,
		Data: FetchResponseData{
			Url:     url,
			Payload: payload,
		},
	}
}

func NewFetcher(ids []uint64, concurrentFetches int, maxIdsPerFetch int) *Fetcher {
	seconds := time.Duration(3) // Approximates the iTunes API rate limit (20 calls/minute)
	t := time.NewTicker(seconds * time.Second)
	logger.Info.Printf("Ticker created, fires every %d seconds\n", seconds)

	f := &Fetcher{
		idPool:            structures.CreatePool[uint64](ids),
		concurrentFetches: concurrentFetches,
		maxIdsPerFetch:    maxIdsPerFetch,

		ticker:       t,
		lastFetchEnd: utils.TimeUnixEpochStart,

		CommandChannel:  make(chan FetcherCommand),
		ResponseChannel: make(chan FetchResponse),
		fetchWaitGroup:  sync.WaitGroup{},
	}

	logger.Info.Printf(
		"Podcast fetcher created with a pool of %d IDs\n",
		len(ids),
	)

	return f
}

func (f *Fetcher) fetch(url string) {
	resp, err := http.Get(url)
	f.fetchWaitGroup.Done()

	statusCode := 500
	if resp != nil {
		statusCode = resp.StatusCode
	}

	if err != nil || statusCode != 200 {
		go func() {
			f.ResponseChannel <- NewFetchResponse(
				false,
				statusCode,
				true,
				url,
				"",
			)
		}()
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		go func() {
			f.ResponseChannel <- NewFetchResponse(
				false,
				statusCode,
				false,
				url,
				"",
			)
		}()
		return
	}

	go func() {
		f.ResponseChannel <- NewFetchResponse(
			true,
			statusCode,
			true,
			url,
			string(body),
		)
	}()
}

func (f *Fetcher) onTick(t time.Time) {
	logger.Info.Println("Pulse at:", t)

	if f.pause {
		logger.Info.Println("Fetcher paused. No actions performed this pulse.")
		return
	}

	tresholdSeconds := 3
	timeSinceFetchEnd := t.Sub(f.lastFetchEnd)
	if timeSinceFetchEnd.Seconds() < float64(tresholdSeconds) {
		logger.Warn.Printf("Request within %d seconds avoided to keep from hitting iTunes rate limits. Pulse ended.\n", tresholdSeconds)
		return
	}

	if f.idPool.Length() == 0 {
		logger.Success.Println("Done crawling IDs")
		os.Exit(0)
	}

	batch := f.idPool.Take(f.concurrentFetches * f.maxIdsPerFetch)

	urls := CreateBatchLookupUrls(
		PODCAST_LOOKUP_URL_BASE,
		batch,
		f.maxIdsPerFetch,
	)
	logger.Info.Printf("Created %d urls from %d ids\n", len(urls), len(batch))

	logger.Info.Printf("Firing %d concurrent requests...\n", len(urls))
	for _, url := range urls {
		f.fetchWaitGroup.Add(1)
		go f.fetch(url)
	}
	logger.Info.Printf("%d concurrent requests fired. Awaiting responses...\n", len(urls))

	f.fetchWaitGroup.Wait()
	logger.Info.Printf("%d concurrent requests completed\n", len(urls))
	f.lastFetchEnd = time.Now()
}

func (f *Fetcher) onCommand(command FetcherCommand) {
	logger.Info.Printf("Command received: %d", command)

	if command == Stop {
		logger.Info.Println("Stop command received. Exitting...")
		os.Exit(0)
	} else if command == Pause {
		logger.Info.Println("Pause command received")
		f.pause = true
	} else if command == Resume {
		logger.Info.Println("Resume command received")
		f.pause = false
	}
}

func (f *Fetcher) Start() {
	go func() {
		logger.Info.Println("Podcast fetcher pulse goroutine created")
		for {
			select {
			case t := <-f.ticker.C:
				logger.System.Println("Running Goroutines:", runtime.NumGoroutine())
				f.onTick(t)
			case command := <-f.CommandChannel:
				f.onCommand(command)
			}
		}
	}()

	logger.Info.Println("Podcast fetcher started")
}
