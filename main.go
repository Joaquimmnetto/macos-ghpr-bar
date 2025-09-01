package main

import (
	"errors"
	"fmt"
	"macos-gh-bar/core"
	"macos-gh-bar/github"
	"macos-gh-bar/native"
	"macos-gh-bar/view"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/progrium/darwinkit/dispatch"
	"github.com/progrium/darwinkit/macos"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

// png, 32x32
var ghPngIcon32x32 = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAYAAABzenr0AAAABHNCSVQICAgIfAhkiAAAAAlwSFlzAAABAwAAAQMB4GlWSgAAABl0RVh0U29mdHdhcmUAd3d3Lmlua3NjYXBlLm9yZ5vuPBoAAAKYSURBVFiFtZdNaxNRFIafM9pQFFdK1UXQ1DRNuyz9A6mFQndFGhAUupN+/IralW66UIQuhSyqG8Gtn6C4EioVsX5EcdWFK2PbFJMeF/cOHce5M5NkeuBwmXvPed/3Zu49cyKqSloTkePAIHARKNgR4DvwzY51VW2lBlXVRAfywAqwDWiCb9vYfCrsBOIy8AhopSAOe8vmlrsSAFSBRhfEYW8A1dQCgD5gNQPisK8CfWkE1I6A3PdarABg/gjJfZ+PFACMA81Q8E17EK8BbzsgeQdcB0rAcmitCYz7vKKqiMgx4CNQ5F+bVNWn/oOIVIEKUA84mNrg+wtVXQ/kXAaehHC/YG5H29/9jGMnpTR3OeEqlxzYM6qKZxUtEW0jjvlOzIWxBOCJyCgw4Qj6lYGAhmN+QkRGPWDWEVBT1ee9sqvqM+CBY3nWwxycKFvulTxgK475QQ/zVQvbLuakZmUfMNcvbAWPw09q0L6q6kFW7KraBj67BJyJWDibFXkC5mkP00iEbUBETmXFbLEGIpbqHrDlyCtnJSAGaytOwGKGAlxYnwDmcHc0QxmU4iHcHdUcwDDQdgT8AMZ6IB+zGFHYbWA4qgl5BbwOPO8Ct4BCB8SXgNs2N7Y58ROKwB+78Bg4gSnRwZ+ujSko9wEvglSAezbmIIZYLVcx3JCsBQJeAp7jfNyN2fliArHva1EdUR7YCQTdsPMLmLL8G3gDXIgRMJKCfIfAf4YwwCSwZwN/Auc7PHQnE8j3MF1WbFc8xWFv+B64CpwDcnaUGAH9MeRNYOq/HAfQNLDvAOrvQsA+MB2ZEwNWATYyELABVJw5Ce9UgCvAZgAwFxOfC8Rt2lznK1O1bXmSiYhg6sKAqt5JiF3AHOCHmgL8L5EXS+d81uVOAAAAAElFTkSuQmCC"

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
	config, err := view.LoadConfiguration(configurationFile)
	if err != nil {
		native.FNSLog("%e", err)
		native.FNSLog("Loading configuration file at %s", defaultConfigurationFile)
		config, err = view.LoadConfiguration(defaultConfigurationFile)
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

func setupStatusBar(app appkit.Application, config view.Configuration) {
	mainMenu := view.NewMenuWithTitle("Open PRs")
	statusItem := appkit.StatusBar_SystemStatusBar().StatusItemWithLength(appkit.VariableStatusItemLength)
	objc.Retain(&statusItem)

	img := view.AppkitImageFromBase64(ghPngIcon32x32)
	statusItem.Button().SetImage(img)
	statusItem.SetMenu(mainMenu)
	statusItem.SetVisible(true)

	refreshTicker := time.NewTicker(config.GithubRefresh())
	go func() {
		for {
			native.NSLog("Refreshing PRs from timer")
			start := time.Now()
			refreshMenuWithPRs(config, app, statusItem, mainMenu)
			native.FNSLog("Refreshed PRs in %s", time.Since(start))
			select {
			case <-refreshTicker.C:
				continue
			}
		}
	}()

}

func refreshMenuWithPRs(config view.Configuration, app appkit.Application, statusItem appkit.StatusItem, mainMenu appkit.Menu) {
	ghops := github.NewGithubOperations(config.GithubToken())
	prsModel, err := core.FetchPRs(ghops, config)
	if err == nil {
		renderStatusMenu(app, statusItem, mainMenu, prsModel, config)
	} else {
		view.DispatchMarkBarButtonOnError(statusItem, errors.Join(err...))
	}
}

func renderStatusMenu(app appkit.Application, statusItem appkit.StatusItem, mainMenu appkit.Menu, prs core.PRMenuModel, config view.Configuration) {
	dispatch.MainQueue().DispatchAsync(func() {
		mainMenu.RemoveAllItems()
		prCount := renderPRs(mainMenu, prs.Shown)
		mainMenu.AddItem(view.MenuSeparator())
		mainMenu.AddItem(view.MenuSeparator())

		if config.RenderHiddenPRs {
			hiddenPRsItem := view.MenuItemNoAction("Hidden PRs", "x")
			hiddenItemsMenu := view.NewMenuWithTitle("Hidden PRs")
			_ = renderPRs(hiddenItemsMenu, prs.Hidden)
			hiddenPRsItem.SetSubmenu(hiddenItemsMenu)
			mainMenu.AddItem(hiddenPRsItem)
		}

		mainMenu.AddItem(view.MenuItem("Refresh", "r", func(sender objc.Object) {
			native.NSLog("Refreshing PRs from button")
			refreshMenuWithPRs(config, app, statusItem, mainMenu)
		}))
		mainMenu.AddItem(view.MenuItem("Quit", "q", func(sender objc.Object) {
			app.Terminate(nil)
		}))
		statusItem.Button().SetTitle(strconv.Itoa(prCount))
	})
}

func renderPRs(menu appkit.Menu, categoryPrs map[string][]github.PullRequest) int {
	renderedCount := 0
	for category, prs := range categoryPrs {
		menu.AddItem(view.MenuSeparator())
		menu.AddItem(view.MenuItemSectionLabel(category))
		menu.AddItem(view.MenuSeparator())
		prsByRepository, sortedRepositories := aggregatePRsByRepository(prs)
		for _, repository := range sortedRepositories {
			prs := prsByRepository[repository]
			if len(prs) == 0 {
				continue
			}
			menu.AddItem(view.MenuItemSubsectionLabel(repository))
			for _, pr := range prs {
				renderedCount = renderedCount + 1
				item := prMenuItem(pr)
				menu.AddItem(item)
			}
		}
		menu.AddItem(view.MenuSeparator())
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
