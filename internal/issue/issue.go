package issue

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-github/v55/github"
)

type Issue struct {
	Title     string
	Body      string
	User      string
	Number    int
	Comments  []Comment
	CreatedAt time.Time
}

type Comment struct {
	CreatedAt time.Time
	Body      string
	User      string
}

type GHClient interface {
	ListByRepo(ctx context.Context, owner string, repo string, options *github.IssueListByRepoOptions) ([]*github.Issue, *github.Response, error)
	ListComments(ctx context.Context, owner string, repo string, number int, options *github.IssueListCommentsOptions) ([]*github.IssueComment, *github.Response, error)
}

type Service struct {
	client GHClient
}

func NewService(client GHClient) *Service {
	return &Service{client: client}
}

func (s *Service) GetIssues(ctx context.Context, owner string, repo string) ([]Issue, error) {
	issues, _, err := s.client.ListByRepo(ctx, owner, repo, &github.IssueListByRepoOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error getting issues: %w", err)
	}

	comments, _, err := s.client.ListComments(ctx, owner, repo, 0, &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error getting replies: %w", err)
	}

	// The only direct link between issue and comments is the issue URL. Use this
	// as the map key to combine the comments to the issue.
	combined := make(map[string]Issue, len(issues))

	for _, issue := range issues {
		combined[issue.GetURL()] = Issue{
			Title:     issue.GetTitle(),
			Body:      issue.GetBody(),
			User:      issue.GetUser().GetLogin(),
			Number:    issue.GetNumber(),
			Comments:  make([]Comment, 0, len(comments)),
			CreatedAt: issue.GetCreatedAt().Time,
		}
	}

	for _, comment := range comments {
		key := comment.GetIssueURL()
		issue, ok := combined[key]
		if !ok {
			fmt.Println("skipping:", key)
			continue
		}

		issue.Comments = append(issue.Comments, Comment{
			CreatedAt: comment.GetCreatedAt().Time,
			Body:      comment.GetBody(),
			User:      comment.GetUser().GetLogin(),
		})

		// Keep the comments in order
		sort.Slice(issue.Comments, func(i, j int) bool {
			return issue.Comments[i].CreatedAt.Before(issue.Comments[j].CreatedAt)
		})

		combined[key] = issue
	}

	out := make([]Issue, 0, len(combined))
	for _, issue := range combined {
		out = append(out, issue)
	}

	return out, nil
}
