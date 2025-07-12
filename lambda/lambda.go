package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type MessageBody struct {
	OrgName  string `json:"orgname"`
	RepoName string `json:"reponame"`
	Branch   string `json:"branchname"`
}

func handler(ctx context.Context, sqsEvent events.SQSEvent) error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN not set in environment")
	}

	for _, message := range sqsEvent.Records {
		var body MessageBody
		if err := json.Unmarshal([]byte(message.Body), &body); err != nil {
			fmt.Printf("Error parsing message body: %v\n", err)
			continue
		}

		workflowFile := "170669863"
		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows/%s/dispatches", body.OrgName, body.RepoName, workflowFile)
		fmt.Printf("API_URL: %s\n", apiURL)

		payload := map[string]interface{}{
			"ref": body.Branch,
			"inputs": map[string]string{
				"triggered_by": "sqs-lambda",
			},
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			fmt.Printf("Error creating JSON payload: %v\n", err)
			continue
		}

		req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonPayload)))
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		req.Header.Set("Authorization", "Bearer "+githubToken)
		req.Header.Set("Accept", "application/vnd.github+json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("❌ HTTP request failed: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Printf("✅ Workflow triggered. Status: %d\n", resp.StatusCode)
		} else {
			fmt.Printf("❌ Failed to trigger workflow. Status: %d\n", resp.StatusCode)
			bodyBytes, _ := io.ReadAll(resp.Body)
			fmt.Printf("Response: %s\n", string(bodyBytes))
		}
	}

	return nil
}

func main() {
	lambda.Start(handler)
}
