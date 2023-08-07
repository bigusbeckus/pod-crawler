package utils

import "time"

// Naive incremental backoff implementation
func IncrementalBackoff(fn func() error) error {
	tresholds := []uint8{
		5,  // 5 seconds
		30, // 30 seconds
		60, // 1 minute
	}

	var err error

	for _, waitTime := range tresholds {
		duration := time.Second * time.Duration(waitTime)
		time.Sleep(duration)

		err = fn()
		if err == nil {
			return nil
		}
	}

	return err
}
