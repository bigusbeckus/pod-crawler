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

func extractIDs(urls []string) []string {
	length := len(urls)

	ids := make([]string, length)
	for i, value := range urls {
		ids[i] = utils.ExtractPodcastId(value)
	}

	logger.Info.Printf("Extracted IDs from %d URLs\n", length)
	return ids
}

func GetIDs() ([]string, error) {
	urls, err := loadInputFile(config.AppConfig.PodcastListFile)
	if err != nil {
		return nil, err
	}

	return extractIDs(urls), nil
}
