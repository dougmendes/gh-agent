package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/dougmendes/gh-agent/utils"
	"github.com/google/go-github/v59/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {
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
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// GET WORKFLOW ID BY FILE NAME
	workflows, _, err := client.Actions.ListWorkflows(ctx, owner, repo, &github.ListOptions{})
	if err != nil {
		panic(err)
	}

	var workflowID int64
	for _, wf := range workflows.Workflows {
		if wf.GetPath() == ".github/workflows/"+workflowFileName {
			workflowID = wf.GetID()
			break
		}
	}

	if workflowID == 0 {
		panic("Workflow not found")
	}

	// GET LAST COMPLETED RUN
	runs, _, err := client.Actions.ListWorkflowRunsByID(ctx, owner, repo, workflowID, &github.ListWorkflowRunsOptions{
		Status: "completed",
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		panic(err)
	}
	if len(runs.WorkflowRuns) == 0 {
		fmt.Println("No completed workflow runs found.")
		os.Exit(0)
	}

	runID := runs.WorkflowRuns[0].GetID()
	fmt.Println("Last run ID:", runID)

	// DOWNLOAD LOGS ZIP
	logsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%d/logs", owner, repo, runID)
	req, _ := http.NewRequest("GET", logsURL, nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("Failed to download logs:", resp.Status)
		os.Exit(1)
	}

	outFile := "logs.zip"
	out, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Println("Logs saved to", outFile)

	// UNZIP LOGS
	if err := utils.Unzip(outFile, "logs/unzipped"); err != nil {
		panic(err)
	}
	fmt.Println("Logs unzipped to logs/unzipped/")
}
