package podcast

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/structures"
)

type FetcherCommand int

const (
	Stop FetcherCommand = iota
	Pause
	Resume
)

type FetcherResponse struct {
	Success bool
	Status  string
	Payload string
}

type Fetcher struct {
	idPool            structures.Pool[uint64]
	concurrentFetches int
	maxIdsPerFetch    int
	ticker            *time.Ticker
	commandChannel    chan FetcherCommand
	responseChannel   chan FetcherResponse
}

func (f *Fetcher) Append(ids ...uint64) {
	f.idPool.Put(ids...)
}

func NewFetcher(ids []uint64, concurrentFetches int, maxIdsPerFetch int) *Fetcher {
	seconds := time.Duration(3) // Approximates the iTunes API rate limit (20 calls/minute)
	t := time.NewTicker(seconds * time.Second)
	logger.Info.Printf("Ticker created, fires every %d seconds\n", seconds)

	f := &Fetcher{
		idPool:            structures.CreatePool[uint64](ids),
		concurrentFetches: concurrentFetches,
		maxIdsPerFetch:    maxIdsPerFetch,
		ticker:            t,
		commandChannel:    make(chan FetcherCommand),
		responseChannel:   make(chan FetcherResponse),
	}

	logger.Info.Printf(
		"Podcast fetcher created with a pool of %d IDs\n",
		len(ids),
	)

	return f
}

func (f *Fetcher) onTick(t time.Time) {
	logger.Info.Println("Pulse at:", t)
	logger.System.Println("Running Goroutines:", runtime.NumGoroutine())

	if t.Second()%2 == 0 {
		r := rand.Int()
		s := r%2 == 0
		var st string
		if s {
			st = "200 OK"
		} else {
			st = "400 ERR"
		}

		f.responseChannel <- FetcherResponse{
			Success: s,
			Status:  st,
			Payload: fmt.Sprint(r),
		}
	}
}

func (f *Fetcher) Start() chan FetcherResponse {
	go func() {
		logger.Info.Println("Podcast fetcher pulse goroutine created")
		for {
			select {
			case t := <-f.ticker.C:
				f.onTick(t)
			}
		}
	}()

	logger.Info.Println("Podcast fetcher started")
	return f.responseChannel
}
