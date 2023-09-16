package output

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mab/issue-download/internal/issue"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func Markdown(issues []issue.Issue, pathPrefix string, owner string, repo string) error {
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
			//if issue.Number == 8 {
			//	println(comment.Body)
			//}
		}

		if err := w.Error(); err != nil {
			return fmt.Errorf("building output file for issue %d: %w", issue.Number, err)
		}

		mdFileName := filepath.Join(pathPrefix, fmt.Sprintf("issue_%d.md", issue.Number))
		f, err := os.Create(mdFileName)
		if err != nil {
			return fmt.Errorf("creating .md output file for issue %d: %w", issue.Number, err)
		}

		// running in a loop so can't defer the close
		if _, err := f.WriteString(buf.String()); err != nil {
			f.Close()
			return fmt.Errorf("writing .md output file for issue %d: %w", issue.Number, err)
		}

		f.Close()

		// read .md file to create HTML
		md, err := ioutil.ReadFile(mdFileName)
		if err != nil {
			log.Fatal(err)
		}

		// replace GitHub URL with local path (there are several variations)
		old1 := "https://user-images.githubusercontent.com/"
		old2 := fmt.Sprintf("https://github.com/%s/%s/assets/", owner, repo)

		new := "./assets/"

		md = bytes.Replace([]byte(md), []byte(old1), []byte(new), -1)
		md = bytes.Replace([]byte(md), []byte(old2), []byte(new), -1)

		html := mdToHTML(md)

		htmlFileName := filepath.Join(pathPrefix, fmt.Sprintf("issue_%d.html", issue.Number))
		htmlF, err := os.Create(htmlFileName)
		if err != nil {
			return fmt.Errorf("creating .html output file for issue %d: %w", issue.Number, err)
		}

		_, err = htmlF.Write(html)
		if err != nil {
			htmlF.Close()
			return fmt.Errorf("writing .html output file for issue %d: %w", issue.Number, err)
		}
		htmlF.Close()
	}

	return nil
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
