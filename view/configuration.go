package view

import (
	"fmt"
	"macos-gh-bar/github"
	"macos-gh-bar/slices"
	"os"
	"regexp"
	"time"

	"github.com/goccy/go-yaml"
)

func LoadConfiguration(configFile string) (Configuration, error) {
	configFileRaw, err := os.ReadFile(configFile)
	if err != nil {
		return Configuration{}, fmt.Errorf("error while reading configuration file %s: %w", configFile, err)
	}
	conf := Configuration{}
	err = yaml.Unmarshal(configFileRaw, &conf)
	if err != nil {
		return Configuration{}, fmt.Errorf("error while parsing configuration file %s: %w", configFile, err)
	}

	return conf, nil
}

type PRFilter struct {
	Category   *string `yaml:"category,omitempty"`
	Repository *string `yaml:"repository,omitempty"`
	Title      *string `yaml:"title,omitempty"`
	Author     *string `yaml:"author,omitempty"`
	Draft      *bool   `yaml:"draft,omitempty"`
}

func (filter PRFilter) Match(pr github.PullRequest) bool {
	components := make([]bool, 0)
	anyMatch := false
	if filter.Repository != nil {
		anyMatch = true
		components = append(components, regexp.MustCompile(*filter.Repository).MatchString(pr.Repository))
	}
	if filter.Title != nil {
		anyMatch = true
		components = append(components, regexp.MustCompile(*filter.Title).MatchString(pr.Title))
	}
	if filter.Author != nil {
		anyMatch = true
		components = append(components, regexp.MustCompile(*filter.Author).MatchString(pr.Author))
	}
	if filter.Draft != nil {
		anyMatch = true
		components = append(components, pr.Draft == *filter.Draft)
	}
	result := anyMatch
	for _, condition := range components {
		result = result && condition
	}
	return result
}

func (filter PRFilter) MatchWithCategory(pr github.PullRequest, category string) bool {
	components := make([]bool, 0)
	if filter.Category != nil {
		components = append(components, regexp.MustCompile(*filter.Category).MatchString(category))
	}
	components = append(components, filter.Match(pr))

	result := true
	for _, condition := range components {
		result = result && condition
	}
	return result
}

type Configuration struct {
	GithubToken           string              `yaml:"github_token"`
	GithubRefreshInterval int                 `yaml:"github_refresh_interval"`
	ShowDrafts            bool                `yaml:"show_drafts"`
	IgnorePRs             []PRFilter          `yaml:"ignore_prs"`
	QueryGroups           map[string][]string `yaml:"query_groups"`
	HidePRs               []PRFilter          `yaml:"hide_prs"`
	RenderHiddenPRs       bool                `yaml:"render_hidden_prs"`
}

func (c Configuration) MatchIgnoredPRs(pr github.PullRequest) bool {
	return slices.Any(c.IgnorePRs, func(filter PRFilter) bool {
		return filter.Match(pr)
	})
}

func (c Configuration) GithubRefresh() time.Duration {
	return time.Duration(c.GithubRefreshInterval) * time.Second
}

func (c Configuration) ResolveGithubToken() string {
	if c.GithubToken == "" {
		return os.Getenv("GH_TOKEN")
	}
	return c.GithubToken
}

func (c Configuration) MatchHidePRs(pr github.PullRequest, category string) bool {
	return slices.Any(c.HidePRs, func(filter PRFilter) bool {
		return filter.MatchWithCategory(pr, category)
	})
}
