package utils

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

func ExtractPodcastId(podcastUrl string) (uint64, error) {
	if !StringIncludes(podcastUrl, "/") {
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
