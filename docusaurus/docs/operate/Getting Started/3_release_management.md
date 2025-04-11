---
sidebar_position: 3
title: PATH Release Management
---

## GitHub Workflow Testing and Release Instructions <!-- omit in toc -->

This document outlines how to test and use the GitHub workflow for building and releasing artifacts for the `buildwithgrove/path` repository.

## Table of Contents <!-- omit in toc -->

- [Testing Workflows Locally](#testing-workflows-locally)
  - [Prerequisites](#prerequisites)
  - [Running the Workflow Locally](#running-the-workflow-locally)
- [Creating and Publishing Releases](#creating-and-publishing-releases)
  - [Versioning](#versioning)
  - [Creating a New Release](#creating-a-new-release)
  - [Manual Workflow Dispatch](#manual-workflow-dispatch)
- [Available Make Commands](#available-make-commands)
- [Troubleshooting](#troubleshooting)

## Testing Workflows Locally

Before pushing your workflow to GitHub, you can test it locally using the `act` tool.

### Prerequisites

1. Install Docker (required for act)
2. Install act:
   - **macOS**: `brew install act`
   - **Linux**: `curl -s https://raw.githubusercontent.com/nektos/act/master/install.sh | sudo bash`
   - **Windows**: `choco install act-cli`

### Running the Workflow Locally

To test the workflow locally:

1. Navigate to your repository root directory
2. Run one of the following commands:

```bash
# Test with default event (push)
act

# Test with a specific event
act workflow_dispatch

# Test with workflow_dispatch and input parameters
act workflow_dispatch -P custom_tag=test

# Test with tag event
act push -e .github/workflows/test-payload.json
```

You can create a test payload file (`.github/workflows/test-payload.json`) with the following content to simulate a tag push:

```json
{
  "ref": "refs/tags/v0.1.0"
}
```

## Creating and Publishing Releases

### Versioning

We follow semantic versioning (MAJOR.MINOR.PATCH):

- MAJOR: breaking changes
- MINOR: new features, no breaking changes
- PATCH: bug fixes

### Creating a New Release

You can use the provided make targets to create a new release:

```bash
# For bug fixes (v1.0.0 -> v1.0.1)
make release_tag_bug_fix

# For minor releases (v1.0.0 -> v1.1.0)
make release_tag_minor_release

# For major releases (manual)
git tag v2.0.0
```

After tagging, push the tag to GitHub:

```bash
git push origin <tag-name>
```

This will automatically trigger the GitHub Action workflow to:

1. Build the binary for multiple platforms
2. Create a GitHub release
3. Upload the built artifacts to the release

### Manual Workflow Dispatch

You can also manually trigger the workflow from GitHub:

1. Go to the Actions tab in your repository
2. Select the "Release artifacts" workflow
3. Click "Run workflow"
4. (Optional) Enter a custom tag suffix
5. Click "Run workflow"

## Available Make Commands

```bash
# Tag a new bug fix release
make release_tag_bug_fix

# Tag a new minor release
make release_tag_minor_release

# Build for local development
make build

# Build release binaries for all supported platforms
make release

# Build and publish a release
make release_publish

# Test the GitHub workflow locally
make test_workflow
```

## Troubleshooting

If you encounter issues with the workflow:

1. Check the GitHub Actions logs for detailed error messages
2. Verify your local environment matches the GitHub Actions environment
3. Run the workflow locally with `act -v` for verbose output

For more help, check the GitHub Actions documentation or open an issue in the repository.
