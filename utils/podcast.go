package utils

import (
	"math"
	"strings"
)

func ExtractPodcastId(podcastUrl string) string {
	if !StringIncludes(podcastUrl, "/") {
		panic("Not a valid url")
	}
	parts := strings.Split(podcastUrl, "/id")
	return parts[1]
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
