package main

import (
	"fmt"
	"macos-gh-bar/github"
	"macos-gh-bar/native"
	"macos-gh-bar/slices"
	"macos-gh-bar/view"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"time"

	yaml "github.com/goccy/go-yaml"
	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

// png, 32x32
var ghPngIcon32x32 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAABAwAAAQMB4GlWSgAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAKYSURBVFiFtZdNaxNRFIafM9pQFFdK1UXQ1DRNuyz9A6mFQndFGhAUupN+/IralW66UIQuhSyqG8Gtn6C4EioVsX5EcdWFK2PbFJMeF/cOHce5M5NkeuBwmXvPed/3Zu49cyKqSloTkePAIHARKNgR4DvwzY51VW2lBlXVRAfywAqwDWiCb9vYfCrsBOIy8AhopSAOe8vmlrsSAFSBRhfEYW8A1dQCgD5gNQPisK8CfWkE1I6A3PdarABg/gjJfZ+PFACMA81Q8E17EK8BbzsgeQdcB0rAcmitCYz7vKKqiMgx4CNQ5F+bVNWn/oOIVIEKUA84mNrg+wtVXQ/kXAaehHC/YG5H29/9jGMnpTR3OeEqlxzYM6qKZxUtEW0jjvlOzIWxBOCJyCgw4Qj6lYGAhmN+QkRGPWDWEVBT1ee9sqvqM+CBY3nWwxycKFvulTxgK475QQ/zVQvbLuakZmUfMNcvbAWPw09q0L6q6kFW7KraBj67BJyJWDibFXkC5mkP00iEbUBETmXFbLEGIpbqHrDlyCtnJSAGaytOwGKGAlxYnwDmcHc0QxmU4iHcHdUcwDDQdgT8AMZ6IB+zGFHYbWA4qgl5BbwOPO8Ct4BCB8SXgNs2N7Y58ROKwB+78Bg4gSnRwZ+ujSko9wEvglSAezbmIIZYLVcx3JCsBQJeAp7jfNyN2fliArHva1EdUR7YCQTdsPMLmLL8G3gDXIgRMJKCfIfAf4YwwCSwZwN/Auc7PHQnE8j3MF1WbFc8xWFv+B64CpwDcnaUGAH9MeRNYOq/HAfQNLDvAOrvQsA+MB2ZEwNWATYyELABVJw5Ce9UgCvAZgAwFxOfC8Rt2lznK1O1bXmSiYhg6sKAqt5JiF3AHOCHmgL8L5EXS+d81uVOAAAAAElFTkSuQmCC"

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

func loadConfiguration(configFile string) (Configuration, error) {
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

func main() {
	selfPath, err := os.Executable()
	if err != nil {
		native.FNSLog("%e", err)
		panic(fmt.Errorf("error while fetching this executable path: %w", err))
	}
	defaultConfigurationFile := filepath.Join(filepath.Dir(selfPath), "config.yml")

	userHome, err := os.UserHomeDir()
	if err != nil {
		native.FNSLog("%e", err)
		panic(fmt.Errorf("error while searching for user home: %w", err))
	}
	configurationFile := filepath.Join(userHome, ".config", "github-bar", "config.yml")

	native.FNSLog("Loading configuration file at %s", configurationFile)
	config, err := loadConfiguration(configurationFile)
	if err != nil {
		native.FNSLog("%e", err)
		native.FNSLog("Loading configuration file at %s", defaultConfigurationFile)
		config, err = loadConfiguration(defaultConfigurationFile)
		if err != nil {
			native.FNSLog("%e", err)
			panic(fmt.Errorf("error while loading default configuration file %s: %w", defaultConfigurationFile, err))
		}
	}

	native.NSLog("Connecting to GitHub API")
	native.NSLog("Booting Application")
	macos.RunApp(func(app appkit.Application, delegate *appkit.ApplicationDelegate) {
		native.NSLog("Starting macOS Menu Bar App")
		app.SetActivationPolicy(appkit.ApplicationActivationPolicyAccessory)
		setupStatusBar(app, config)
		native.NSLog("Status bar set up successfully")
	})
}

func setupStatusBar(app appkit.Application, config Configuration) {
	ghops := github.NewGithubOperations(config.GithubToken())
	menu := appkit.NewMenuWithTitle("Open PRs")
	objc.Retain(&menu)
	statusItem := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(appkit.VariableStatusItemLength)
	objc.Retain(&statusItem)

	img := view.AppkitImageFromBase64(ghPngIcon32x32)
	objc.Retain(&img)
	statusItem.Button().SetImage(img)
	statusItem.SetMenu(menu)
	statusItem.SetVisible(true)

	refreshTicker := time.NewTicker(config.GithubRefresh())
	go func() {
		for {
			native.NSLog("Refreshing PRs")
			var createdPRs = make([]github.PullRequest, 0)
			var reviewerPRs = make([]github.PullRequest, 0)
			var prs []github.PullRequest
			var err error
			if config.SearchCreatedPRs {
				createdPRs, err = ghops.CreatedOpenPRs()
				if err != nil {
					native.FNSLog("Error searching user created PRs: %v", err)
					goto endCurrentLoop
				}
			}
			if config.SearchReviewerPRs {
				reviewerPRs, err = ghops.ReviewerOpenPRs()
				if err != nil {
					native.FNSLog("Error searching PRs tagging user as reviewer: %v", err)
					goto endCurrentLoop
				}
			}
			prs = append(createdPRs, reviewerPRs...)
			for _, query := range config.ExtraQueries {
				queriedPRs, err := ghops.SearchIssues(query)
				if err != nil {
					native.FNSLog("Error searching PRs matching query %s: %v", query, err)
					goto endCurrentLoop
				}
				prs = append(prs, queriedPRs...)
			}
			renderStatusMenu(app, statusItem, menu, prs, config)
		endCurrentLoop:
			if err != nil {
				view.DispatchMarkBarButtonOnError(statusItem, err)
			}
			select {
			case <-refreshTicker.C:
				continue
			}
		}
	}()

}

func renderStatusMenu(app appkit.Application, statusItem appkit.StatusItem, menu appkit.Menu, prs []github.PullRequest, config Configuration) {
	dispatch.MainQueue().DispatchAsync(func() {
		menu.RemoveAllItems()
		prCount := renderPRs(menu, prs, config)
		native.FNSLog("Rendered %d out of %d PRs", prCount, len(prs))
		menu.AddItem(view.MenuSeparator())
		menu.AddItem(view.MenuSeparator())
		menu.AddItem(view.MenuItem("Quit", "q", func(sender objc.Object) {
			app.Terminate(nil)
		}))
		statusItem.Button().SetTitle(strconv.Itoa(prCount))
	})

}

func renderPRs(menu appkit.Menu, allPRs []github.PullRequest, config Configuration) int {
	prsByRepository, sortedRepositories := aggregatePRsByRepository(allPRs)
	renderedCount := 0
	for _, repository := range sortedRepositories {
		prs := prsByRepository[repository]
		prs = slices.Filter(prs, func(pr github.PullRequest) bool {
			if !config.ShowDrafts && pr.Draft {
				return false
			}
			ignore := slices.Any(config.IgnoreRepositoriesRegexes(), func(regex *regexp.Regexp) bool { return regex.MatchString(pr.Repository) }) ||
				slices.Any(config.IgnorePRRegexes(), func(regex *regexp.Regexp) bool { return regex.MatchString(pr.Title) })
			if ignore {
				return false
			}
			return true
		})
		if len(prs) == 0 {
			continue
		}
		menu.AddItem(view.MenuSeparator())
		repositoryLabel := appkit.NewMenuItem()
		repositoryLabel.SetEnabled(false)
		repositoryLabel.SetView(view.SubsectionTitleLabel(repository))
		objc.Retain(&repositoryLabel)
		menu.AddItem(repositoryLabel)
		for _, pr := range prs {
			renderedCount = renderedCount + 1
			item := prMenuItem(pr)
			menu.AddItem(item)
		}
	}
	return renderedCount
}

func prMenuItem(pr github.PullRequest) appkit.MenuItem {
	return view.MenuItem(fmt.Sprintf("%s [#%d]", pr.Title, pr.Number), "",
		func(sender objc.Object) {
			err := exec.Command("open", pr.URL).Start()
			view.DispatchAlertOnError(err)
		},
	)
}

func aggregatePRsByRepository(prs []github.PullRequest) (map[string][]github.PullRequest, []string) {
	repositories := make([]string, 0)
	aggregated := make(map[string][]github.PullRequest)
	for _, pr := range prs {
		aggregated[pr.Repository] = append(aggregated[pr.Repository], pr)
	}
	for repository, _ := range aggregated {
		repositories = append(repositories, repository)
	}
	sort.Strings(repositories)
	return aggregated, repositories
}
