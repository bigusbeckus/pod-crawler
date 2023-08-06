package app

import (
	"os"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/models"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database/service"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/podcast"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/structures"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
	// "gorm.io/gorm/clause"
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
	db, _ := database.GetInstance()
	tx := db.Begin()

	resultsCount := o.payloads.Length()
	logger.Info.Printf("Converting %d results to models\n", resultsCount)
	results := o.payloads.Take(resultsCount)
	podcasts := make([]models.Podcast, 0, resultsCount)
	for _, result := range results {
		p, err := service.PodcastFromItunesResult(result)
		if err != nil {
			tx.Rollback()
			logger.Error.Fatalf("Save failed: Unable to convert results into database models: %v\n", err)
		}
		podcasts = append(podcasts, *p)
	}

	tx.CreateInBatches(podcasts, 1000)
	if tx.Error != nil {
		logger.Error.Printf(
			"Save failed: Unable to save results to database: %v\nPausing further fetches and retrying...\n",
			tx.Error,
		)
		o.fetcher.CommandChannel <- podcast.Pause

		err := utils.IncrementalBackoff(func() error {
			tx.CreateInBatches(podcasts, 1000)
			return tx.Error
		})

		if err != nil {
			tx.Rollback()
			logger.Error.Fatalln("Save failed: Unable to save results to database with incremental backoff")
		}
		o.fetcher.CommandChannel <- podcast.Resume
	}

	tx.Commit()
	if tx.Error != nil {
		logger.Error.Fatalf("Failed to save %d results to database: %v\n", resultsCount, tx.Error)
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
	ids := podcast.ExtractLookupIDs(url)
	if err != nil {
		o.Fail(ids) // TODO: Find a way to do individual validation on result entries
	}

	failures := make([]uint64, 0, p.ResultCount)
	successes := make([]podcast.ItunesResult, 0, p.ResultCount)
	for _, result := range p.Results {
		isCollectionNameEmpty := result.CollectionName == nil || len(strings.TrimSpace(*result.CollectionName)) == 0
		if isCollectionNameEmpty {
			failures = append(failures, uint64(result.CollectionId))
			continue
		}

		isCollectionCensoredNameEmpty := result.CollectionCensoredName == nil || len(strings.TrimSpace(*result.CollectionCensoredName)) == 0
		if isCollectionCensoredNameEmpty {
			result.CollectionCensoredName = result.CollectionName
		}

		successes = append(successes, result)
	}

	if len(failures) > 0 {
		go o.Fail(failures)
	}

	if len(successes) > 0 {
		logger.Success.Printf("Parsed %d results, %d total\n", len(successes), o.payloads.Length())
		go o.Succeed(successes)
	}

	go o.handleUnfetched(ids, successes)
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
