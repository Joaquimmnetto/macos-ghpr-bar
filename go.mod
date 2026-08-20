go 1.24

module macos-gh-bar

require (
	github.com/goccy/go-yaml v1.18.0
	github.com/google/go-github/v74 v74.0.0
	github.com/progrium/darwinkit v0.5.0
)

require github.com/google/go-querystring v1.1.0 // indirect

replace github.com/progrium/darwinkit => github.com/pmoust/darwinkit v0.5.1-0.20260729123318-0a368ccd4ca9
