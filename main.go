package main

import (
	"fmt"
	"os"

	"github.com/bigusbeckus/podcast-feed-fetcher/internal/app"
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

	fmt.Println("Initialization complete")
}

func SetupDB() {
	// Run database migrations
	logger.Info.Println("Database migrations started")
	err := database.RunMigrations()
	if err != nil {
		logger.Error.Fatalf("\n%v\n", err.Error())
	}
	logger.Info.Println("Database migrations successful")
}

func main() {
	Init()
	logger.PrintHeading("Podcast Feed Fetcher")

	SetupDB()

	app.Start(config.AppConfig.SaveTreshold)
}
