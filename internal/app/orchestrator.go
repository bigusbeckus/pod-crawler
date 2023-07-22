package app

import (
	"log"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
)

func Start() {
	ids, err := podcast.GetIDs()
	if err != nil {
		logger.Error.Fatalf("Failed to get podcast IDs from input: %v", err)
	}

	f := podcast.NewFetcher(
		ids,
		config.AppConfig.ConcurrentFetchBatchSize,
		config.AppConfig.SingleFetchIDsCount,
	)

	responseChannel := f.Start()

	for {
		select {
		case r := <-responseChannel:
			var currentLogger *log.Logger
			if r.Success {
				currentLogger = logger.Success
			} else {
				currentLogger = logger.Error
			}

			currentLogger.Printf("Status: %s\n", r.Status)
			logger.Info.Printf(r.Payload)
		}
	}
}
