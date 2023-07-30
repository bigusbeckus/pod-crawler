package app

import (
	"log"
	"os"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
)

func Start() {
	ids, err := podcast.GetIDs()
	if err != nil {
		logger.Error.Fatalf("Failed to get podcast IDs from input: %v", err)
	}

	ids, err = filterCrawled(ids)
	if err != nil {
		logger.Error.Fatalf("Failed to filter out crawled podcasts: %v", err)
	}

	if len(ids) == 0 {
		logger.Success.Println("All IDs have already been processed. No further action is needed")
		os.Exit(0)
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

			currentLogger.Printf("Status: %d\n", r.Status)
		}
	}
}

func filterCrawled(ids []uint64) ([]uint64, error) {
	db, err := database.GetInstance()
	if err != nil {
		return nil, err
	}

	logger.Info.Println("Looking up known ids from database")
	var crawledIds []uint64
	err = db.Model(&models.Podcast{}).Pluck("ItunesID", &crawledIds).Error
	if err != nil {
		return nil, err
	}
	logger.Info.Printf("Lookup successful. Found %d entries\n", len(crawledIds))

	logger.Info.Println("Comparing slices and extracting unprocessed IDs")
	unprocessedIds := utils.LeftDiff(ids, crawledIds)
	logger.Info.Printf("Comparison done. %d unprocessed IDs found\n", len(unprocessedIds))

	return unprocessedIds, nil
}
