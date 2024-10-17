package main

import (
	"context"
	"dagger/gh/internal/dagger"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Work with GitHub pull requests.
func (m *Gh) PullRequest() *PullRequest {
	return &PullRequest{Gh: m}
}

type PullRequest struct {
	// +private
	Gh *Gh
}

// Create a pull request on GitHub.
func (m *PullRequest) Create(
	ctx context.Context,

	// Assign people by their login. Use "@me" to self-assign.
	//
	// +optional
	assignees []string,

	// The branch into which you want your code merged.
	//
	// +optional
	base string,

	// Body for the pull request.
	//
	// +optional
	body string,

	// Read body text from file.
	//
	// +optional
	bodyFile *dagger.File,

	// Mark pull request as a draft.
	//
	// +optional
	draft bool,

	// Use commit info for title and body. (Requires repository source)
	//
	// +optional
	fill bool,

	// Use first commit info for title and body. (Requires repository source)
	//
	// +optional
	fillFirst bool,

	// Use commits msg+body for description. (Requires repository source)
	//
	// +optional
	fillVerbose bool,

	// The branch that contains commits for your pull request (default [current branch], required when no repository source is available).
	//
	// +optional
	head string,

	// Add labels by name.
	//
	// +optional
	labels []string,

	// Add the pull request to a milestone by name.
	//
	// +optional
	milestone string,

	// Disable maintainer's ability to modify pull request.
	//
	// +optional
	noMaintainerEdit bool,

	// Add the pull request to projects by name.
	//
	// +optional
	projects []string,

	// Request reviews from people or teams by their handle.
	//
	// +optional
	reviewers []string,

	// Template file to use as starting body text.
	//
	// +optional
	template *dagger.File,

	// Title for the pull request.
	//
	// +optional
	title string,

	// GitHub token.
	//
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) error {
	if m.Gh.Source == nil {
		if head == "" {
			return errors.New("\"head\" is required when no git repository is available")
		}

		if fill || fillFirst || fillVerbose {
			return errors.New("\"fill\", \"fillFirst\" and \"fillVerbose\" require a git repository source")
		}
	}

	if !(fill || fillFirst || fillVerbose) && title == "" {
		return errors.New("\"title\" is required when none of the fill options are configured")
	}

	ctr := m.Gh.container(token, repo)

	args := []string{"gh", "pr", "create"}

	for _, assignee := range assignees {
		args = append(args, "--assignee", assignee)
	}

	if base != "" {
		args = append(args, "--base", base)
	}

	if body != "" {
		args = append(args, "--body", body)
	}

	if bodyFile != nil {
		ctr.WithMountedFile("/work/tmp/body", bodyFile)
		args = append(args, "--body-file", "/work/tmp/body")
	}

	if draft {
		args = append(args, "--draft")
	}

	if fill {
		args = append(args, "--fill")
	}

	if fillFirst {
		args = append(args, "--fill-first")
	}

	if fillVerbose {
		args = append(args, "--fill-verbose")
	}

	if head != "" {
		args = append(args, "--head", head)
	}

	for _, label := range labels {
		args = append(args, "--label", label)
	}

	if milestone != "" {
		args = append(args, "--milestone", milestone)
	}

	if noMaintainerEdit {
		args = append(args, "--no-maintainer-edit")
	}

	for _, project := range projects {
		args = append(args, "--project", project)
	}

	for _, reviewer := range reviewers {
		args = append(args, "--reviewer", reviewer)
	}

	if template != nil {
		ctr.WithMountedFile("/work/tmp/template", template)
		args = append(args, "--template", "/work/tmp/template")
	}

	if title != "" {
		args = append(args, "--title", title)
	}

	_, err := ctr.WithExec(args).Sync(ctx)

	return err
}

// Close a pull request on GitHub.
func (m *PullRequest) Close(
	ctx context.Context,

	// Pull request number to close.
	//
	// +optional
	pullRequest string,

	// Add a comment when closing the pull request.
	//
	// +optional
	comment string,

	// Delete the local and remote branch after closing.
	//
	// +optional
	deleteBranch bool,

	// GitHub token.
	//
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) error {
	ctr := m.Gh.container(token, repo)

	args := []string{"gh", "pr", "close", pullRequest}

	if comment != "" {
		args = append(args, "--comment", comment)
	}

	if deleteBranch {
		args = append(args, "--delete-branch")
	}

	_, err := ctr.WithExec(args).Sync(ctx)

	return err
}

// Add a review to a pull request.
func (m *PullRequest) Review(
	// Pull request number, url or branch name.
	pullRequest string,

	// Specify the body of a review.
	//
	// +optional
	body string,

	// Read body text from file.
	//
	// +optional
	bodyFile *dagger.File,
) *PullRequestReview {
	return &PullRequestReview{
		PullRequest: pullRequest,
		Body:        body,
		BodyFile:    bodyFile,
		Gh:          m.Gh,
	}
}

func (m *PullRequest) List(
	ctx context.Context,

	// Filter by pull request state: {open|closed|merged|all}.
	//
	// +optional
	state string,

	// Filter by pull request base branch.
	//
	// +optional
	base string,

	// Filter by head branch.
	//
	// +optional
	head string,

	// Filter by head branch using regex pattern.
	//
	// +optional
	headRegex string,

	// GitHub token.
	//
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) (string, error) {
	ctr := m.Gh.container(token, repo)

	args := []string{"gh", "pr", "list", "--json", "number,headRefName", "--limit", "1000"}

	if state != "" {
		args = append(args, "--state", state)
	}

	if base != "" {
		args = append(args, "--base", base)
	}

	if head != "" {
		args = append(args, "--head", head)
	}

	output, err := ctr.WithExec(args).Stdout(ctx)
	if err != nil {
		return "", err
	}

	var prList []struct {
		Number      int    `json:"number"`
		HeadRefName string `json:"headRefName"`
	}
	if err := json.Unmarshal([]byte(output), &prList); err != nil {
		return "", fmt.Errorf("failed to parse PR list: %w", err)
	}

	if headRegex != "" {
		regex, err := regexp.Compile(headRegex)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}

		for _, pr := range prList {
			if regex.MatchString(pr.HeadRefName) {
				return strconv.Itoa(pr.Number), nil
			}
		}
		return "", fmt.Errorf("no pull requests found matching the regex pattern")
	}

	if len(prList) == 0 {
		return "", fmt.Errorf("no pull requests found")
	}

	return strconv.Itoa(prList[0].Number), nil
}

// Update an existing pull request on GitHub.
func (m *PullRequest) Update(
	ctx context.Context,

	// Pull request number to update.
	//
	// +optional
	pullRequest string,

	// Assign people by their login. Use "@me" to self-assign.
	//
	// +optional
	assignees []string,

	// The branch into which you want your code merged.
	//
	// +optional
	base string,

	// Body for the pull request.
	//
	// +optional
	body string,

	// Read body text from file.
	//
	// +optional
	bodyFile *dagger.File,

	// Add labels by name.
	//
	// +optional
	labels []string,

	// Add the pull request to a milestone by name.
	//
	// +optional
	milestone string,

	// Add the pull request to projects by name.
	//
	// +optional
	projects []string,

	// Request reviews from people or teams by their handle.
	//
	// +optional
	reviewers []string,

	// Title for the pull request.
	//
	// +optional
	title string,

	// GitHub token.
	//
	// +optional
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	//
	// +optional
	repo string,
) error {
	ctr := m.Gh.container(token, repo)

	args := []string{"gh", "pr", "edit", pullRequest}

	for _, assignee := range assignees {
		args = append(args, "--add-assignee", assignee)
	}

	if base != "" {
		args = append(args, "--base", base)
	}

	if body != "" {
		args = append(args, "--body", body)
	}

	if bodyFile != nil {
		ctr = ctr.WithMountedFile("/work/tmp/body", bodyFile)
		args = append(args, "--body-file", "/work/tmp/body")
	}

	for _, label := range labels {
		args = append(args, "--add-label", label)
	}

	if milestone != "" {
		args = append(args, "--milestone", milestone)
	}

	for _, project := range projects {
		args = append(args, "--add-project", project)
	}

	for _, reviewer := range reviewers {
		args = append(args, "--add-reviewer", reviewer)
	}

	if title != "" {
		args = append(args, "--title", title)
	}

	_, err := ctr.WithExec(args).Sync(ctx)

	return err
}

// TODO: revisit if these should be private
type PullRequestReview struct {
	// +private
	PullRequest string

	// +private
	Body string

	// +private
	BodyFile *dagger.File

	// +private
	Gh *Gh
}

// Approve a pull request.
func (m *PullRequestReview) Approve(ctx context.Context) error {
	return m.do(ctx, "approve")
}

// Comment on a pull request.
func (m *PullRequestReview) Comment(ctx context.Context) error {
	return m.do(ctx, "comment")
}

// Request changes on a pull request.
func (m *PullRequestReview) RequestChanges(ctx context.Context) error {
	return m.do(ctx, "request-changes")
}

// Request changes on a pull request.
func (m *PullRequestReview) do(ctx context.Context, action string) error {
	args := []string{"gh", "pr", "review", m.PullRequest, "--" + action}

	_, err := m.Gh.container(nil, "").
		With(func(c *dagger.Container) *dagger.Container {
			if m.Body != "" {
				args = append(args, "--body", m.Body)
			}

			if m.BodyFile != nil {
				c = c.WithMountedFile("/work/tmp/body", m.BodyFile)
				args = append(args, "--body-file", "/work/tmp/body")
			}

			return c
		}).
		WithExec(args).
		Sync(ctx)

	return err
}
