# GhBar

macOS menu bar app that lists your open GitHub pull requests. It runs as an accessory app (no Dock icon): click the GitHub mark in the menu bar to browse PRs grouped by your config, open them in the browser, refresh on a timer, or quit.

## Requirements

- **macOS** (Apple Silicon or Intel)
- **Go 1.24+** — run from source
- **Xcode Command Line Tools** — CGO builds and app icons (`clang`, `iconutil`, `sips`)
- **JDK + Gradle wrapper** — build the `.app` bundle and DMG (included in the repo)

## Configuration

GhBar loads config in this order:

1. `~/.config/github-bar/config.yml` (preferred for installed apps)
2. `config.yml` next to the executable (e.g. inside `GhBar.app/Contents/MacOS/`)

GitHub authentication:

- Set `github_token` in YAML, or
- Leave it empty and export **`GH_TOKEN`** (recommended so the token is not stored in the bundle)

Example `config.yml` (see the repo file for a full sample):

```yaml
github_refresh_interval: 120
query_groups:
  "To Review":
    - "is:pr is:open review-requested:@me archived:false"
  "Created":
    - "is:pr is:open author:@me archived:false"
ignore_prs:
  - author: 'dependabot.*$'
  - draft: true
```

`query_groups` defines menu sections and GitHub search queries. `ignore_prs` / `hide_prs` use regex filters on title, author, repository, draft, and category.

## Run from source

```bash
export GH_TOKEN=ghp_…   # if not in config
go run .
```

Logs go to the macOS unified log (Console.app, filter by process name).

## Build and install the app

Gradle tasks chain: `buildGoApp` → `buildAppIcon` → `createApkFolder` → optional `signApp` / `createDmg`.

| Task | Purpose |
|------|---------|
| `./gradlew buildGoApp` | Compile Go binary to `./app` |
| `./gradlew createApkFolder` | Assemble `GhBar.app` in the project root |
| `./gradlew installApk` | Copy `GhBar.app` to `~/Applications/` |
| `./gradlew signApp` | Code-sign the bundle (hardened runtime + entitlements) |
| `./gradlew createDmg` | Produce `GhBar.dmg` |

### Architecture

Default build is a **universal** binary (arm64 + x86_64):

```bash
./gradlew buildGoApp                    # universal (default)
./gradlew buildGoApp -PgoArch=arm64     # Apple Silicon only
./gradlew buildGoApp -PgoArch=amd64     # Intel only
```

### Version and bundle ID

Set in [`gradle.properties`](gradle.properties) (overridable with `-P`):

- `ghBarVersion` — marketing version (`CFBundleShortVersionString`)
- `ghBarBuildVersion` — build number (`CFBundleVersion`)
- `ghBarBundleIdentifier` — e.g. `com.joaquimmnetto.ghbar`

### Code signing

Ad-hoc signing is used if you do not pass an identity:

```bash
./gradlew signApp -PsigningIdentity='Developer ID Application: Your Name (TEAMID)'
```

Distribution to other Macs also requires Apple notarization; see [`PACKAGING_DIAGNOSTIC.md`](PACKAGING_DIAGNOSTIC.md) for remaining packaging gaps.

## Project layout

- `main.go`, `core/`, `github/`, `view/` — application logic and UI
- `native/` — small CGO bridge to Foundation logging
- `Info.plist`, `entitlements.plist`, `assets/AppIcon.png` — bundle metadata and icon source
- `build.gradle` — Go build, icon, app bundle, DMG, install
