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

type fetchResponseData struct {
	Url     string
	Payload string
}

type fetchResponse struct {
	Success     bool
	IsBodyValid bool
	Status      int
	Data        fetchResponseData
}

type fetcher struct {
	idPool            structures.Pool[uint64]
	concurrentFetches int
	maxIdsPerFetch    int

	ticker          *time.Ticker
	lastFetchEnd    time.Time
	commandChannel  chan FetcherCommand
	responseChannel chan fetchResponse
	fetchWaitGroup  sync.WaitGroup

	pause bool
}

func (f *fetcher) Append(ids ...uint64) {
	f.idPool.Put(ids...)
}

func NewFetchResponse(
	success bool,
	status int,
	isBodyValid bool,
	url string,
	payload string,
) fetchResponse {
	return fetchResponse{
		Success:     success,
		IsBodyValid: isBodyValid,
		Status:      status,
		Data: fetchResponseData{
			Url:     url,
			Payload: payload,
		},
	}
}

func NewFetcher(ids []uint64, concurrentFetches int, maxIdsPerFetch int) *fetcher {
	seconds := time.Duration(3) // Approximates the iTunes API rate limit (20 calls/minute)
	t := time.NewTicker(seconds * time.Second)
	logger.Info.Printf("Ticker created, fires every %d seconds\n", seconds)

	f := &fetcher{
		idPool:            structures.CreatePool[uint64](ids),
		concurrentFetches: concurrentFetches,
		maxIdsPerFetch:    maxIdsPerFetch,

		ticker:       t,
		lastFetchEnd: utils.TimeUnixEpochStart,

		commandChannel:  make(chan FetcherCommand),
		responseChannel: make(chan fetchResponse),
		fetchWaitGroup:  sync.WaitGroup{},
	}

	logger.Info.Printf(
		"Podcast fetcher created with a pool of %d IDs\n",
		len(ids),
	)

	return f
}

func (f *fetcher) fetch(url string) {
	// defer f.fetchWaitGroup.Done()

	resp, err := http.Get(url)
	f.fetchWaitGroup.Done()
	logger.Info.Println("Done fetch")

	if err != nil || resp.StatusCode != 200 {
		go func() {
			f.responseChannel <- NewFetchResponse(
				false,
				resp.StatusCode,
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
			f.responseChannel <- NewFetchResponse(
				false,
				resp.StatusCode,
				false,
				url,
				"",
			)
		}()
		return
	}

	go func() {
		f.responseChannel <- NewFetchResponse(
			true,
			resp.StatusCode,
			true,
			url,
			string(body),
		)
	}()
}

func (f *fetcher) onTick(t time.Time) {
	logger.Info.Println("Pulse at:", t)

	tresholdSeconds := 3
	timeSinceFetchEnd := t.Sub(f.lastFetchEnd)
	if timeSinceFetchEnd.Seconds() < float64(tresholdSeconds) {
		logger.Warn.Printf("Request within %d seconds avoided to keep from hitting iTunes rate limits. Pulse ended.\n", tresholdSeconds)
		return
	}

	if f.pause {
		logger.Info.Println("Fetcher paused. No actions performed this pulse.")
		return
	}

	if f.idPool.Length() == 0 {
		logger.Success.Println("Out of ids")
		return
	}

	batch := f.idPool.Take(f.concurrentFetches * f.maxIdsPerFetch)

	logger.Info.Printf("Preparing %d ids for fetch\n", len(batch))
	urls := CreateBatchLookupUrls(
		PODCAST_LOOKUP_URL_BASE,
		batch,
		f.maxIdsPerFetch,
	)
	logger.Info.Println("IDs prepared")

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

func (f *fetcher) onCommand(command FetcherCommand) {
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

func (f *fetcher) Start() chan fetchResponse {
	go func() {
		logger.Info.Println("Podcast fetcher pulse goroutine created")
		for {
			select {
			case t := <-f.ticker.C:
				logger.System.Println("Running Goroutines:", runtime.NumGoroutine())
				f.onTick(t)
			case command := <-f.commandChannel:
				f.onCommand(command)
			}
		}
	}()

	logger.Info.Println("Podcast fetcher started")
	return f.responseChannel
}
