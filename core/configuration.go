package core

import (
	"fmt"
	"github.com/goccy/go-yaml"
	"os"
	"regexp"
	"time"
)

type Configuration struct {
	githubToken           string   `yaml:"github_token,omitempty"`
	GithubRefreshInterval int      `yaml:"github_refresh_interval"`
	ShowDrafts            bool     `yaml:"show_drafts"`
	IgnoreRepositories    []string `yaml:"ignore_repositories"`
	IgnorePRs             []string `yaml:"ignore_prs"`
	ExtraQueries          []string `yaml:"extra_queries"`
	SearchCreatedPRs      bool     `yaml:"search_created_prs"`
	SearchReviewerPRs     bool     `yaml:"search_reviewer_prs"`

	ignoreRepositoriesRegexes []*regexp.Regexp
	ignorePRsRegexes          []*regexp.Regexp
}

func stringsToRegexes(strings []string) []*regexp.Regexp {
	regexes := make([]*regexp.Regexp, 0, len(strings))
	for _, ignoreRegexStr := range strings {
		regexes = append(regexes, regexp.MustCompile(ignoreRegexStr))
	}
	return regexes
}

func (c Configuration) IgnoreRepositoriesRegexes() []*regexp.Regexp {
	if c.ignoreRepositoriesRegexes == nil {
		c.ignoreRepositoriesRegexes = stringsToRegexes(c.IgnoreRepositories)
	}
	return c.ignoreRepositoriesRegexes
}

func (c Configuration) IgnorePRRegexes() []*regexp.Regexp {
	if c.ignorePRsRegexes == nil {
		c.ignorePRsRegexes = stringsToRegexes(c.IgnorePRs)
	}
	return c.ignorePRsRegexes
}

func (c Configuration) GithubRefresh() time.Duration {
	return time.Duration(c.GithubRefreshInterval) * time.Second
}

func (c Configuration) GithubToken() string {
	if c.githubToken == "" {
		return os.Getenv("GH_TOKEN")
	}
	return c.githubToken
}

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
