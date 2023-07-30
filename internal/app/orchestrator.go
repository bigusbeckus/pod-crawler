package app

import (
	"os"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/service"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/structures"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
)

type orchestrator struct {
	saveTreshold int
	payloads     structures.Pool[podcast.ItunesResult]
	failedIds    structures.Pool[uint64]
	fetcher      *podcast.Fetcher
}

func newOrchestrator(saveTreshold int) orchestrator {
	o := orchestrator{
		saveTreshold: saveTreshold,
		payloads:     structures.CreatePool([]podcast.ItunesResult{}),
		failedIds:    structures.CreatePool([]uint64{}),
	}
	logger.Info.Printf("Orchestrator created with a save treshold of %d results\n", saveTreshold)

	return o
}

func Start(saveTreshold int) {
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

	o := newOrchestrator(saveTreshold)
	o.fetcher = podcast.NewFetcher(
		ids,
		config.AppConfig.ConcurrentFetchBatchSize,
		config.AppConfig.SingleFetchIDsCount,
	)

	o.fetcher.Start()

	for {
		select {
		case r := <-o.fetcher.ResponseChannel:
			o.onFetchResponse(r)
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

func (o *orchestrator) Save() {
	db, err := database.GetInstance()
	if err != nil {
		logger.Error.Println("Save failed: Unable to get database instance")
		o.fetcher.CommandChannel <- podcast.Pause

		utils.IncrementalBackoff(func() error {
			db, err = database.GetInstance()
			return err
		})
		if err != nil {
			logger.Error.Fatalln("Save failed: Unable to get database instance with incremental backoff")
		}
		o.fetcher.CommandChannel <- podcast.Resume
	}

	resultsCount := o.payloads.Length()
	logger.Info.Printf("Converting %d results to models\n", resultsCount)
	results := o.payloads.Take(resultsCount)
	podcasts := make([]models.Podcast, resultsCount)
	for _, result := range results {
		p, err := service.PodcastFromItunesResult(result)
		if err != nil {
			o.fetcher.CommandChannel <- podcast.Stop
			logger.Error.Fatalln("Save failed: Unable to convert results into database models")
		}
		podcasts = append(podcasts, *p)
	}

	tx := db.Save(podcasts)
	if tx.Error != nil {
		logger.Error.Println("Save failed: Unable to save results to database")
		o.fetcher.CommandChannel <- podcast.Pause

		utils.IncrementalBackoff(func() error {
			tx = db.Save(podcasts)
			return tx.Error
		})

		if tx.Error != nil {
			logger.Error.Fatalln("Save failed: Unable to save results to database with incremental backoff")
		}
		o.fetcher.CommandChannel <- podcast.Resume
	}

	logger.Success.Printf("Successfully saved %d/%d results to database\n", tx.RowsAffected, resultsCount)
}

func (o *orchestrator) Succeed(results []podcast.ItunesResult) {
	o.payloads.Put(results...)
	payloadsCount := o.payloads.Length()
	if payloadsCount > o.saveTreshold {
		logger.Info.Printf("Saving %d results to database...\n", payloadsCount)
		o.Save()
	}
}

func (o *orchestrator) Fail(ids []uint64) {
	o.failedIds.Put(ids...)
}

func (o *orchestrator) Requeue(ids []uint64) {
	o.fetcher.Append(ids...)
	o.fetcher.Shuffle()
}

func (o *orchestrator) onFetchResponse(msg podcast.FetchResponse) {
	if !msg.Success {
		go o.onFetchFail(msg.IsBodyValid, msg.Data.Url)
		return
	}

	go o.onFetchSuccess(msg.Data.Url, msg.Data.Payload)
}

func (o *orchestrator) onFetchSuccess(url string, payload string) {
	p, err := podcast.ParseLookupResponse(payload)
	// Success
	if err == nil {
		o.Succeed(p.Results)
	}

	ids := podcast.ExtractLookupIDs(url)
	if err != nil {
		// Failure
		o.Fail(ids)
	} else {
		// Success but requires ids
		go o.handleUnfetched(ids, p.Results)
	}
}

func (o *orchestrator) onFetchFail(isBodyValid bool, url string) {
	failedIds := podcast.ExtractLookupIDs(url)
	if !isBodyValid {
		logger.Error.Printf(
			"Fetch failed for %d IDs. Entries will not be requeued due to malformed response bodies",
			len(failedIds),
		)
		o.Fail(failedIds)
		return
	}

	logger.Error.Printf(
		"Fetch failed for %d IDs. Entries requeued",
		len(failedIds),
	)
	o.Requeue(failedIds)
}

func (o *orchestrator) handleUnfetched(ids []uint64, results []podcast.ItunesResult) {
	resultIds := make([]uint64, len(results))
	for i := range results {
		resultIds[i] = uint64(results[i].CollectionId)
	}

	unfetchedIds := utils.LeftDiff(ids, resultIds)
	logger.Info.Printf("Requeued %d ids that were missing from result", len(unfetchedIds))
	o.Requeue(unfetchedIds)
}
