package asset

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
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

	authNeeded, err := requiresAuth(url)
	if err != nil {
		return fmt.Errorf("determining if auth required for image download: %w", err)
	}

	if authNeeded {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}

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

	urlPath, err := assetPath(url)
	if err != nil {
		return fmt.Errorf("processing %s: %w", url, err)
	}

	urlPath = strings.TrimPrefix(urlPath, "/"+pathPrefix)

	filename := filepath.Base(urlPath) + extension
	dirname := filepath.Dir(urlPath)

	if !strings.Contains(dirname, "assets") {
		dirname = filepath.Join("assets", dirname)
	}

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

// github haved changed the way that authorization is required
// for image upload. There is an auth failure if we try to
// get images from the older endpoints and provide an authorization field.
// https://github.blog/changelog/2023-05-09-more-secure-private-attachments/
func requiresAuth(ghURL string) (bool, error) {
	u, err := url.Parse(ghURL)
	if err != nil {
		return false, fmt.Errorf("parsing url: %w", err)
	}

	switch u.Hostname() {
	case "user-images.githubusercontent.com":
		return false, nil
	default:
		return true, nil
	}
}

func assetPath(assetURL string) (string, error) {
	u, err := url.Parse(assetURL)
	if err != nil {
		return "", fmt.Errorf("parsing url: %w", err)
	}

	return u.Path, nil
}
