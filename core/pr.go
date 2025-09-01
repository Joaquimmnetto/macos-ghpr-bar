package core

import (
	"fmt"
	"macos-gh-bar/github"
	"macos-gh-bar/native"
	"macos-gh-bar/slices"
	"macos-gh-bar/view"
	"time"
)

type PRMenuModel struct {
	Hidden map[string][]github.PullRequest
	Shown  map[string][]github.PullRequest
}

func FetchPRs(ghops *github.GhOperations, config view.Configuration) (PRMenuModel, []error) {
	var searchErrors []error
	prs := slices.MapParallelMany(config.QueryGroups, func(category string, queries []string) []github.PullRequest {
		categoryPRs := slices.ParallelMany(queries, func(query string) []github.PullRequest {
			start := time.Now()
			queriedPRs, err := ghops.SearchIssues(query)
			native.FNSLog("Ran Github query %s in %s", query, time.Since(start))
			queriedPRs = slices.Filter(queriedPRs, func(pr github.PullRequest) bool {
				return config.MatchIgnoredPRs(pr) == false
			})
			if err != nil {
				searchErrors = append(searchErrors, fmt.Errorf("error searching PRs matching query %s: %w", query, err))
			}
			return queriedPRs
		})
		return categoryPRs
	})
	/*
		var categoryWg sync.WaitGroup
		for category, queries := range config.QueryGroups {
		categoryWg.Add(1)
		go func() {
			defer categoryWg.Done()
			queryPRs := slices.ParallelMany(queries, func(query string) []github.PullRequest {
				start := time.Now()
				queriedPRs, err := ghops.SearchIssues(query)
				native.FNSLog("Ran Github query %s in %s", query, time.Since(start))
				queriedPRs = slices.Filter(queriedPRs, func(pr github.PullRequest) bool {
					return config.MatchIgnoredPRs(pr) == false
				})
				if err != nil {
					searchErrors = append(searchErrors, fmt.Errorf("error searching PRs matching query %s: %w", query, err))
				}
				return queriedPRs
			})
			prs[category] = queryPRs
		}()

		categoryWg.Wait()
	*/
	prsToHide := make(map[string][]github.PullRequest, len(prs))
	prsToShow := make(map[string][]github.PullRequest, len(prs))
	for category, queries := range prs {
		prsToShow[category] = make([]github.PullRequest, 0)
		toHide, toShow := slices.Split(queries, func(pr github.PullRequest) bool {
			return config.MatchHidePRs(pr, category)
		})
		prsToHide[category] = append(prsToHide[category], toHide...)
		prsToShow[category] = append(prsToShow[category], toShow...)
	}
	return PRMenuModel{
		Hidden: prsToHide,
		Shown:  prsToShow,
	}, searchErrors
}
