package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mab/issue-download/internal/issue"
)

func Markdown(issues []issue.Issue, pathPrefix string) error {
	if err := os.MkdirAll(pathPrefix, 0o755); err != nil {
		return fmt.Errorf("making output directory: %w", err)
	}

	for _, issue := range issues {
		var buf strings.Builder

		w := writer{w: &buf}

		w.WriteString("***\n")
		w.WriteString(fmt.Sprintf("# ISSUE %d %s %s ", issue.Number, issue.User, issue.CreatedAt.Format(time.RFC3339)) + "\n")
		w.WriteString((issue.Title + "\n"))
		w.WriteString("***\n")

		w.WriteString(issue.Body + "\n\n")

		for _, comment := range issue.Comments {
			w.WriteString("***\n")
			w.WriteString(fmt.Sprintf("# REPLY %s %s ", comment.User, comment.CreatedAt.Format(time.RFC3339)) + "\n")
			w.WriteString("***\n")
			w.WriteString(comment.Body + "\n\n")
		}

		if err := w.Error(); err != nil {
			return fmt.Errorf("building output file for issue %d: %w", issue.Number, err)
		}

		f, err := os.Create(filepath.Join(pathPrefix, fmt.Sprintf("issue_%d.md", issue.Number)))
		if err != nil {
			return fmt.Errorf("creating output file for issue %d: %w", issue.Number, err)
		}

		// running in a loop so can't defer the close

		if _, err := f.WriteString(buf.String()); err != nil {
			f.Close()
			return fmt.Errorf("writing output file for issue %d: %w", issue.Number, err)
		}

		f.Close()
	}

	return nil
}
