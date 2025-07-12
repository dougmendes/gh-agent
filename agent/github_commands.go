package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/dougmendes/gh-agent/utils"
	"github.com/google/go-github/v59/github"
)

func getWorkflowIDByFileName(ctx context.Context, client *github.Client, owner, repo, workflowFileName string) (int64, error) {
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
	return runID, nil
}

func downloadLogsZip(owner, repo, token string, runID int64) (string, error) {
	logsURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%d/logs", owner, repo, runID)
	req, _ := http.NewRequest("GET", logsURL, nil)
	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Failed to download logs:", resp.StatusCode)
	}

	outFile := "logs.zip"
	out, err := os.Create(outFile)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	fmt.Println("Logs saved to", outFile)

	return outFile, nil

}

func unzipLogs(outFile string) error {
	return utils.Unzip(outFile, "logs/unzipped")
}
