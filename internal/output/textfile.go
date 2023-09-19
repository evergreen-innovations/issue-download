package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/evergreen-innovations/issue-download/internal/issue"
)

func TextFile(issues []issue.Issue, pathPrefix string) error {
	if err := os.MkdirAll(pathPrefix, 0o755); err != nil {
		return fmt.Errorf("making output directory: %w", err)
	}

	for _, issue := range issues {
		f, err := os.Create(filepath.Join(pathPrefix, fmt.Sprintf("issue_%d.txt", issue.Number)))
		if err != nil {
			return fmt.Errorf("creating output file for issue %d: %w", issue.Number, err)
		}
		// running a loop so can't defer the file close!

		//
		w := writer{w: f}
		w.WriteString(strings.Repeat("=", 10) + fmt.Sprintf(" ISSUE %d %s %s ", issue.Number, issue.User, issue.CreatedAt.Format(time.RFC3339)) + strings.Repeat("=", 10) + "\n")
		w.WriteString(issue.Title + "\n")

		w.WriteString(strings.Repeat("=", 60) + "\n")

		w.WriteString(issue.Body + "\n\n")

		for _, comment := range issue.Comments {
			w.WriteString(strings.Repeat("=", 10) + fmt.Sprintf(" REPLY %s %s ", comment.User, comment.CreatedAt.Format(time.RFC3339)) + strings.Repeat("=", 10) + "\n")
			w.WriteString(comment.Body + "\n\n")
		}

		if err := w.Error(); err != nil {
			f.Close()
			return fmt.Errorf("writing output file for issue %d: %w", issue.Number, err)
		}

		f.Close()
	}

	return nil
}
