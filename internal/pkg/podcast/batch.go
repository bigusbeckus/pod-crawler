package podcast

import "math"

func extractBatch(ids *[]string, batchSize int) []string {
	endIndex := int(math.Min(float64(batchSize), float64(len(*ids))))
	currentBatch := (*ids)[:endIndex]
	*ids = (*ids)[endIndex:]
	return currentBatch
}
