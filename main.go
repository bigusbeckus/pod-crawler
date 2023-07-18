package main

import (
	"fmt"
	"os"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/config"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/database"
	"github.com/bigusbeckus/podcast-feed-fetcher/internal/pkg/logger"
)

func Init() {
	fmt.Println("Performing initialization tasks")

	// Load config
	fmt.Print("Loading config file...")
	err := config.Load()
	if err != nil {
		fmt.Printf("\n%v\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Done")

	// Initialize logger
	fmt.Print("Initializing logger...")
	err = logger.Init()
	if err != nil {
		fmt.Printf("\n%v\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Done")

	// Run database migrations
	fmt.Print("Running database migrations...")
	err = database.RunMigrations()
	if err != nil {
		fmt.Printf("\n%v\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("Done")

	fmt.Println("Initialization complete")
}

func main() {
	Init()

	println()
	fmt.Println("======================")
	fmt.Println("Podcast Feed Fetcher")
	fmt.Println("======================")

}
