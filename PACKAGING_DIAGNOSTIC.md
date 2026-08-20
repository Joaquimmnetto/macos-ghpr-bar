# DMG Packaging Diagnostic

Snapshot of what's still needed to publish GhBar as a proper, installable macOS DMG menu-bar
app. Written 2026-08-19, revisit when picking this work back up.

The gradle build already has a solid skeleton: `buildGoApp` -> `createApkFolder` -> `signApp`
-> `createDmg` -> `installApk`. The gaps below are what's missing on top of that skeleton.

## Code signing / Gatekeeper

- `signApp` defaults to ad-hoc signing (`-`). Real distribution needs an actual **Developer ID
  Application** certificate and a paid Apple Developer Program membership, passed via
  `-PsigningIdentity`.
- No **notarization** step at all (`xcrun notarytool submit` + `xcrun stapler staple`). Without
  this, Gatekeeper will block/warn on any machine that isn't yours, even with a real cert.
- No `codesign --verify --deep --strict` or `spctl -a -vvv` sanity check after signing.
- `createDmg` never signs or notarizes the DMG itself, and runs on the *unsigned* `GhBar.app` in
  the project dir rather than depending on `signApp`. Correct order should be: build -> sign ->
  notarize app -> staple -> build dmg (from the stapled app) -> optionally sign/notarize the dmg
  too.

## Architecture

- **Addressed:** `buildGoApp` supports `-PgoArch=universal|arm64|amd64` (default `universal`).
  Per-arch CGO builds write to `build/go/GhBar-{arm64,amd64}`; universal mode merges with
  `lipo -create` into `./app`. When the host CPU differs from the target, cross-build sets
  `CGO_*FLAGS` with `clang -arch` (requires Xcode CLI tools).
- Single-arch modes copy one intermediate binary to `./app` for local iteration; release builds
  should use the default universal target.

## Info.plist gaps

- **Addressed:** Bundle metadata is driven from [`gradle.properties`](gradle.properties)
  (`ghBarBundleIdentifier`, `ghBarVersion`, `ghBarBuildVersion`) and expanded into
  [`Info.plist`](Info.plist) when `createApkFolder` runs. Includes `CFBundleExecutable`,
  `CFBundleShortVersionString`, `CFBundleIconFile`, and `AppIcon.icns` built from
  [`assets/AppIcon.png`](assets/AppIcon.png) (same glyph as the menu-bar PNG in `main.go`).
- Bump `ghBarVersion` for user-visible releases; increment `ghBarBuildVersion` for each shipped
  build (required if you add Sparkle or similar update checks later).
- `LSUIElement` remains set (menu-bar-only) — no change needed.

## DMG polish (cosmetic, optional)

- No custom background image, icon layout, or window size/position (AppleScript-driven
  `.DS_Store`, or a tool like `create-dmg`). Currently just app + `Applications` symlink dropped
  in a folder — functional but plain.

## Distribution / config concerns

- `GithubToken` can live in plaintext `config.yml` shipped inside the bundle, and falls back to
  the `GH_TOKEN` env var. Fine for personal use, but if this DMG is meant for other people,
  there's no onboarding/setup flow for them to provide their own token securely (e.g. via
  Keychain) — currently it's BYO-config-file.
- `entitlements.plist` only grants `network.client`, no hardened-runtime extras — consistent
  with what notarization needs, so no gap there, just noting it's new/untracked and should get
  committed.

## Docs / process

- **Addressed:** [`README.md`](README.md) covers config, local dev, Gradle build/install, and signing basics.
- No CI/release automation (tag -> build -> sign -> notarize -> publish DMG to GitHub
  Releases) — purely manual right now via local gradle tasks.

## Bottom line

The packaging *mechanics* (bundle, entitlements, DMG creation, install task) are already there
and reasonably good. What's actually blocking a "proper," shareable DMG is **notarization + a
real Developer ID cert**. Architecture and bundle metadata (id, version, icon) are in place;
confirm with `./gradlew createApkFolder` and `plutil -p GhBar.app/Contents/Info.plist`.
