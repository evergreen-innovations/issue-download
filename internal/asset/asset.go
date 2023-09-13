package asset

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"mab/issue-download/internal/issue"
)

type Service struct {
	client *http.Client
	token  string
}

func NewService(client *http.Client, githubToken string) *Service {
	return &Service{client: client, token: githubToken}
}

func (s *Service) DownloadImages(issues []issue.Issue, pathPrefix string) error {
	for _, issue := range issues {
		if err := s.downloadImages(issue.Body, pathPrefix); err != nil {
			return fmt.Errorf("issue %d: %w", issue.Number, err)
		}

		for _, comment := range issue.Comments {
			if err := s.downloadImages(comment.Body, pathPrefix); err != nil {
				return fmt.Errorf("issue %d: %w", issue.Number, err)
			}
		}
	}

	return nil
}

func (s *Service) downloadImages(body string, pathPrefix string) error {
	for _, asset := range extractImages(body) {
		fmt.Println("asset:", asset)
		if err := s.downloadAsset(extractAssetURL(asset), pathPrefix); err != nil {
			return fmt.Errorf("downloading asset for %s: %w", asset, err)
		}
	}

	return nil
}

func (s *Service) downloadAsset(url, pathPrefix string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	var body bytes.Buffer
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %s", body.String())
	}

	if resp.StatusCode != http.StatusOK {
		if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
			// Don't want to print reams of html
			body.Reset()
		}
		return fmt.Errorf("got status code: %d, %s", resp.StatusCode, body.String())
	}

	extension := ""
	switch resp.Header.Get("Content-Type") {
	case "image/png":
		extension = ".png"
	}

	urlPath := strings.TrimPrefix(url, "https://github.com")
	filename := filepath.Base(urlPath) + extension
	dirname := filepath.Dir(urlPath)

	if pathPrefix == "" {
		pathPrefix = "."
	}

	fullpath := filepath.Join(pathPrefix, dirname)

	if err := os.MkdirAll(fullpath, 0o750); err != nil {
		return fmt.Errorf("making directory: %w", err)
	}

	f, err := os.Create(filepath.Join(fullpath, filename))
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	_, err = body.WriteTo(f)
	if err != nil {
		return fmt.Errorf("writing to file: %w", err)
	}

	return nil
}

var findImage = regexp.MustCompile(`\!\[[a-z0-9_\-\.]*\]\(https://.*\)`)

func extractImages(str string) []string {
	return findImage.FindAllString(str, -1)
}

var findAssetURL = regexp.MustCompile(`\((https://.*)\)`)

func extractAssetURL(str string) string {
	matches := findAssetURL.FindStringSubmatch(str)
	if len(matches) == 0 {
		return ""
	}

	return matches[1]
}
