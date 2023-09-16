package issue

import (
	"context"
	"fmt"
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
	IssueURL  string
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

	// NOTE by JS: This could likely be much simpler (no need for combined if we obtain comments for each issue). To be cleaned up.
	combined := make(map[string]Issue, len(issues))

	for _, issue := range issues {

		// get comments for EACH issue; otherwise, limit of 100 max comments is hit very quickly for larger project
		commentsIssue, _, err := s.client.ListComments(ctx, owner, repo, issue.GetNumber(), &github.IssueListCommentsOptions{
			ListOptions: github.ListOptions{
				PerPage: 100,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("error getting replies: %w", err)
		}

		var comments []Comment
		for _, commentIssue := range commentsIssue {
			comments = append(comments, Comment{
				IssueURL:  commentIssue.GetIssueURL(),
				CreatedAt: commentIssue.GetCreatedAt().Time,
				Body:      commentIssue.GetBody(),
				User:      commentIssue.GetUser().GetLogin(),
			})
		}

		combined[issue.GetURL()] = Issue{
			Title:     issue.GetTitle(),
			Body:      issue.GetBody(),
			User:      issue.GetUser().GetLogin(),
			Number:    issue.GetNumber(),
			Comments:  comments,
			CreatedAt: issue.GetCreatedAt().Time,
		}
	}

	out := make([]Issue, 0, len(combined))
	for _, issue := range combined {
		out = append(out, issue)
	}

	return out, nil
}
