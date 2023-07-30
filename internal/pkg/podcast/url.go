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

const PODCAST_LOOKUP_URL_BASE = "https://itunes.apple.com/lookup?entity=podcast&id="

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

func CreateBatchLookupUrls(baseUrl string, podcastIds []uint64, idsPerUrl int) []string {
	urlsCount := int(math.Ceil(float64(len(podcastIds)) / float64(idsPerUrl)))
	urls := make([]string, urlsCount)

	for i := range urls {
		startIndex := i * idsPerUrl
		endIndex := int(math.Min(float64(startIndex)+float64(idsPerUrl), float64(len(podcastIds))))

		currentBatch := podcastIds[startIndex:endIndex]
		urls[i] = baseUrl + utils.JoinNumbers(currentBatch, ",") // strings.Join(currentBatch, ",")
	}

	return urls
}

func ExtractLookupIDs(url string) []uint64 {
	parts := strings.Split(url, "&id=")
	idParams := parts[1]
	idStrs := strings.Split(idParams, ",")

	ids := make([]uint64, len(idStrs))
	for _, id := range idStrs {
		idNum, err := strconv.ParseUint(id, 0, 0)
		if err != nil {
			idNum = 0
		}
		ids = append(ids, idNum)
	}

	return ids
}
