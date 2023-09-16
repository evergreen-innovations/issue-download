package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/go-github/v55/github"

	"mab/issue-download/internal/issue"
	"mab/issue-download/internal/output"
)

func main() {
	var mainErr error
	defer func() {
		if mainErr != nil {
			fmt.Println("error:", mainErr)
			os.Exit(1)
		}
	}()

	args := os.Args[1:]

	if len(args) != 2 {
		fmt.Println("usage: issue-download owner repo")
		os.Exit(1)
	}

	owner := args[0]
	repo := args[1]

	GH_TOKEN, ok := os.LookupEnv("GH_TOKEN")
	if !ok {
		mainErr = errors.New("must supply GH_TOKEN envirnoment variable")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Seems a shame that we have 2 http clients but passing the httpClient
	// here seems to modify the first http client which prevents it from
	// making calls later.
	client := github.NewClient(nil).WithAuthToken(GH_TOKEN)

	issueService := issue.NewService(client.Issues)

	/*httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	assetService := asset.NewService(httpClient, GH_TOKEN)*/

	issues, err := issueService.GetIssues(ctx, owner, repo)
	if err != nil {
		mainErr = err
		return
	}

	outputdir := filepath.Join(owner, repo)
	/*if err := assetService.DownloadImages(issues, outputdir); err != nil {
		mainErr = err
		return
	}*/

	// TODO add prefix directory
	if err := os.MkdirAll(filepath.Join(owner, repo), 0o755); err != nil {
		mainErr = fmt.Errorf("making output directory: %w", err)
		return
	}

	if err := output.Markdown(issues, outputdir); err != nil {
		mainErr = fmt.Errorf("writing output: %w", err)
		return
	}
}
