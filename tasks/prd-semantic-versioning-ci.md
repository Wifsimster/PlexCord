# PRD: Automated Semantic Versioning CI/CD Pipeline

## Introduction

Implement automated semantic versioning for PlexCord using Conventional Commits. The pipeline will automatically determine version bumps based on commit messages, update version files, create Git tags, and trigger releases on every successful push to `main`. This removes manual version management and ensures consistent, traceable releases.

## Goals

- Automate version determination based on Conventional Commits (feat: = minor, fix: = patch, BREAKING CHANGE: = major)
- Automatically release on every push to `main` after all CI checks pass
- Keep version files in sync across the codebase (`frontend/package.json`, Go ldflags)
- Generate changelogs automatically from commit history
- Maintain backward compatibility with existing release workflow for manual overrides

## User Stories

### US-001: Configure Conventional Commits Validation
**Description:** As a developer, I want commit messages validated against Conventional Commits spec so that the automation can reliably determine version bumps.

**Acceptance Criteria:**
- [ ] Add commitlint configuration file (`.commitlintrc.json` or similar)
- [ ] Configure rules for Conventional Commits format (`type(scope): description`)
- [ ] Supported types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`
- [ ] Document commit message format in CONTRIBUTING.md or README

### US-002: Create Semantic Release Workflow
**Description:** As a maintainer, I want a GitHub Action that automatically determines the next version and creates a release so that I don't have to manually tag releases.

**Acceptance Criteria:**
- [ ] Create new workflow file `.github/workflows/release-please.yml` (or semantic-release equivalent)
- [ ] Workflow triggers on push to `main` branch
- [ ] Workflow only runs after CI workflow completes successfully
- [ ] Version bump determined by commit messages since last release:
  - `fix:` commits trigger PATCH bump (e.g., 4.3.0 → 4.3.1)
  - `feat:` commits trigger MINOR bump (e.g., 4.3.0 → 4.4.0)
  - `BREAKING CHANGE:` in commit body or `!` after type triggers MAJOR bump (e.g., 4.3.0 → 5.0.0)
- [ ] Commits with only `docs:`, `style:`, `test:`, `ci:`, `chore:` do NOT trigger a release

### US-003: Update Version Files Automatically
**Description:** As a developer, I want version files updated automatically so that the codebase reflects the current version without manual edits.

**Acceptance Criteria:**
- [ ] `frontend/package.json` version field updated to new version
- [ ] `CHANGELOG.md` updated with new release section
- [ ] Version injected into Go binary via `-ldflags` during release build (already exists, ensure it uses the new tag)
- [ ] Changes committed by automation bot with appropriate commit message
- [ ] Version updates do not trigger recursive releases

### US-004: Generate Changelog Automatically
**Description:** As a user, I want release notes generated from commit history so that I can see what changed in each version.

**Acceptance Criteria:**
- [ ] Changelog grouped by type (Features, Bug Fixes, Breaking Changes, etc.)
- [ ] Each entry includes commit message and short SHA
- [ ] Breaking changes highlighted prominently
- [ ] Changelog included in GitHub Release body
- [ ] `CHANGELOG.md` file in repo root updated with each release
- [ ] CHANGELOG.md follows Keep a Changelog format (https://keepachangelog.com)

### US-005: Gate Releases on CI Success
**Description:** As a maintainer, I want releases blocked until all CI checks pass so that broken code is never released.

**Acceptance Criteria:**
- [ ] Release workflow uses `workflow_run` trigger or `needs` dependency on CI
- [ ] Release only proceeds if lint-go, lint-frontend, test-go, and build jobs all pass
- [ ] Failed CI prevents any version bump or release creation
- [ ] Security scan failures logged but don't block release (continue-on-error already set)

### US-006: Create Git Tag and GitHub Release
**Description:** As a user, I want each version to have a Git tag and GitHub Release so that I can download specific versions.

**Acceptance Criteria:**
- [ ] Git tag created in format `vX.Y.Z` (e.g., `v4.4.0`)
- [ ] GitHub Release created with:
  - Title: `PlexCord vX.Y.Z`
  - Body: Auto-generated changelog
  - Assets: Built binaries for Windows, macOS, Linux (from existing release workflow)
- [ ] Release is not marked as draft or prerelease (unless it's a prerelease version)

### US-007: Trigger Existing Build Pipeline on New Tags
**Description:** As a maintainer, I want the existing release build workflow to trigger automatically when a new version tag is created so that binaries are built and attached to the release.

**Acceptance Criteria:**
- [ ] Existing `release.yml` workflow continues to work unchanged
- [ ] New semantic versioning workflow creates tags that trigger `release.yml`
- [ ] Binaries attached to the correct GitHub Release
- [ ] No duplicate releases or builds

## Functional Requirements

- FR-1: The system must parse Conventional Commits to determine version bump type
- FR-2: The system must skip releases for commits that don't warrant a version bump (docs, style, test, ci, chore only)
- FR-3: The system must update `frontend/package.json` version field before creating a release
- FR-4: The system must create a Git tag in `vX.Y.Z` format
- FR-5: The system must generate a changelog grouped by commit type
- FR-6: The system must wait for CI workflow to complete successfully before releasing
- FR-7: The system must handle first release (no previous tags) gracefully
- FR-8: The system must not create releases for merge commits from release automation (avoid infinite loops)
- FR-9: The system must update `CHANGELOG.md` in repo root with each release

## Non-Goals

- No pre-release versioning (alpha, beta, rc) - all releases are production
- No monorepo support (single version for entire project)
- No manual version override in workflow (use existing manual tag workflow if needed)
- No npm publish (this is a desktop app, not an npm package)
- No branch-specific versioning (only `main` branch releases)

## Technical Considerations

- **Tool Choice**: Use [Release Please](https://github.com/google-github-actions/release-please-action) (Google's tool) or [semantic-release](https://github.com/semantic-release/semantic-release). Release Please is simpler and creates a "Release PR" pattern; semantic-release is more automated.
- **Existing Workflow Integration**: The current `release.yml` triggers on `v*.*.*` tags. The new automation should create these tags to leverage existing build infrastructure.
- **Version File Sync**: `frontend/package.json` needs file update; Go version is injected via ldflags at build time (no file change needed).
- **CHANGELOG.md**: Maintain a `CHANGELOG.md` in repo root following Keep a Changelog format. Updated automatically with each release.
- **GitHub Token**: Release automation needs `contents: write` permission to create tags and releases.
- **Avoid Loops**: Commits made by the release bot should be ignored to prevent infinite release cycles.

## Success Metrics

- Every merge to `main` with releasable commits produces a versioned release within 15 minutes
- Zero manual intervention required for standard releases
- Version in `frontend/package.json` always matches the latest Git tag
- All releases include platform binaries (Windows, macOS, Linux)

## Open Questions

1. Should we use Release Please (creates Release PRs for review) or semantic-release (fully automatic)?
   - **Recommendation**: semantic-release for true continuous delivery per user's choice of "2A"

## Decisions Made

- Continue versioning from current `4.3.0` - first automated release will be `4.3.1`, `4.4.0`, or `5.0.0` depending on commits
- Include `CHANGELOG.md` in repo root (Keep a Changelog format)
