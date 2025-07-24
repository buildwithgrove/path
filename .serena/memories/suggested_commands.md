# PATH Development Commands

## Building and Running

- `make path_build` - Build the PATH binary locally
- `make path_run` - Run PATH as a standalone binary (requires CONFIG_PATH)
- `make path_up` - Start local Tilt development environment with dependencies
- `make path_down` - Tear down local Tilt development environment

## Testing

- `make test_unit` - Run all unit tests (`go test ./... -short -count=1`)
- `make test_all` - Run unit tests plus E2E tests for key services
- `make e2e_test SERVICE_IDS` - Run E2E tests for specific Shannon service IDs (e.g., `make e2e_test eth,poly`)
- `make load_test SERVICE_IDS` - Run Shannon load tests
- `make go_lint` - Run Go linters (`golangci-lint run --timeout 5m --build-tags test`)

## Configuration

- `make config_prepare_shannon_e2e` - Prepare Shannon E2E configuration

## Release and Versioning

- `make release_build_cross` - Build binaries for multiple platforms with version info
- `make release_build_local` - Build binary for current platform only
- `make release_tag_dev` - Tag a new dev release
- `make release_tag_bug_fix` - Tag a new bug fix release
- `make release_tag_minor_release` - Tag a new minor release

## System Commands (Darwin)

- `git` for version control
- `grep` for text search (use `rg` ripgrep if available)
- `find` for file searching
- `ls` for directory listing
- `cd` for directory navigation
