package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dougmendes/gh-agent/auth"
	"github.com/joho/godotenv"
)

func main() {
	ctx := context.Background()
	// CONFIG
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	owner := os.Getenv("OWNER")
	repo := os.Getenv("REPO")
	workflowFileName := os.Getenv("WORKFLOWFILENAME")
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		fmt.Println("GITHUB_TOKEN is not set")
		os.Exit(1)
	}

	// AUTH
	githubClient := auth.GithubAuth(token, ctx)
	// GET WORKFLOW ID BY FILE NAME
	runID, err := getWorkflowIDByFileName(ctx, githubClient, owner, repo, workflowFileName)

	// DOWNLOAD LOGS ZIP
	outFile, err := downloadLogsZip(owner, repo, token, runID)
	if err != nil {
		log.Fatalf("Error downloading logs: %v", err)
	}

	err = unzipLogs(outFile)
	if err != nil {
		log.Fatalf("Error unzipping logs: %v", err)
	}
	fmt.Println("Logs unzipped to logs/unzipped/")
	fmt.Println("Done!")
}
