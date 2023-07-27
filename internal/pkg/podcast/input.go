package podcast

import (
	"os"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/utils"
)

func loadInputFile(filename string) ([]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	contentLines := strings.Split(strings.TrimSpace(string(content)), "\n")

	logger.Info.Printf("Loaded %d URLs from file: `%s`\n", len(contentLines), filename)
	return contentLines, nil
}

func extractIDs(urls []string) []uint64 {
	length := len(urls)

	ids := make([]uint64, length)
	for i, value := range urls {
		id, err := utils.ExtractPodcastId(value)
		if err != nil {
			logger.Error.Print(err)
			continue
		}
		ids[i] = id
	}

	logger.Info.Printf("Extracted %d IDs from %d URLs\n", len(ids), length)
	return ids
}

func GetIDs() ([]uint64, error) {
	urls, err := loadInputFile(config.AppConfig.PodcastListFile)
	if err != nil {
		return nil, err
	}

	return extractIDs(urls), nil
}
