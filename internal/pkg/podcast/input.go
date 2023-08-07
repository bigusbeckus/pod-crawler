package podcast

import (
	"os"
	"strings"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
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

func GetIDs() ([]uint64, error) {
	urls, err := loadInputFile(config.AppConfig.PodcastListFile)
	if err != nil {
		return nil, err
	}

	return extractIDs(urls), nil
}
