package podcast

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
)

func parseUrl(podcastUrl string) (uint64, error) {
	if !utils.StringIncludes(podcastUrl, "/") {
		return 0, errors.New(
			fmt.Sprintf("invalid podcast url: %s", podcastUrl),
		)
	}

	parts := strings.Split(podcastUrl, "/id")

	id, err := strconv.ParseUint(parts[len(parts)-1], 0, 0)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Extracts ids given a list of podcast urls
func extractIDs(urls []string) []uint64 {
	length := len(urls)

	ids := make([]uint64, length)
	for i, value := range urls {
		id, err := parseUrl(value)
		if err != nil {
			logger.Warn.Print(err)
			continue
		}
		ids[i] = id
	}

	logger.Info.Printf("Extracted %d IDs from %d URLs\n", len(ids), length)
	return ids
}

func CreateBatchLookupLinks(baseUrl string, podcastIds []string, batchSize int) []string {
	batchLinksCount := int(math.Ceil(float64(len(podcastIds)) / float64(batchSize)))
	batchLinks := make([]string, batchLinksCount)

	for i := range batchLinks {
		startIndex := i * batchSize
		endIndex := int(math.Min(float64(startIndex)+float64(batchSize), float64(len(podcastIds))))

		currentBatch := podcastIds[startIndex:endIndex]
		batchLinks[i] = baseUrl + strings.Join(currentBatch, ",")
	}

	return batchLinks
}
