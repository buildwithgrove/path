---
sidebar_position: 4
title: PATH Release Management
---

## GitHub Workflow Testing and Release Instructions

This document outlines how to test and use the GitHub workflow for building and releasing artifacts for the Path project.

## Table of Contents

- [GitHub Workflow Testing and Release Instructions](#github-workflow-testing-and-release-instructions)
- [Table of Contents](#table-of-contents)
- [Building and Releasing](#building-and-releasing)
  - [Local Development Build](#local-development-build)
  - [Creating Release Artifacts](#creating-release-artifacts)
- [Versioning and Tagging](#versioning-and-tagging)
  - [Creating a New Release Tag](#creating-a-new-release-tag)
- [Testing Workflows Locally](#testing-workflows-locally)
  - [Prerequisites](#prerequisites)
  - [Setting Up Secrets](#setting-up-secrets)
  - [Testing Specific Workflows](#testing-specific-workflows)
- [Available Make Commands](#available-make-commands)

## Building and Releasing

### Local Development Build

To build the binary for local development:

```bash
make path_build
```

This will create the binary in the `bin` directory.

### Creating Release Artifacts

To build release binaries for all supported platforms:

```bash
make path_release
```

This command builds binaries for the following platforms:

- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64

The release artifacts will be created in the `release` directory as compressed `tar.gz` files.

## Versioning and Tagging

We follow semantic versioning: `MAJOR.MINOR.PATCH`

### Creating a New Release Tag

You can create a new release tag using the following command:

```bash
make path_release_tag
```

This will:

1. Prompt you to specify the type of release: `bug`, `minor`, or `major`
2. Generate the appropriate new version number based on the latest existing tag
3. Create a new git tag with the version number

After creating the tag, push it to GitHub:

```bash
git push origin <tag_name>
```

This will trigger the release workflow to build and publish the release artifacts.

## Testing Workflows Locally

Before pushing your changes to GitHub, you can test the workflows locally using the `act` tool.

### Prerequisites

Install the `act` tool for local GitHub Actions testing:

```bash
make install_act
```

This will install `act` using Homebrew on macOS or the installation script on Linux.

### Setting Up Secrets

Create a `.secrets` file in the repository root with your GitHub token:

```bash
GITHUB_TOKEN=your_github_token
```

You can create a token at: [github.com/settings/tokens](ttps://github.com/settings/tokens)

### Testing Specific Workflows

To test the build and push workflow:

```bash
make workflow_test_build_and_push
```

To test the release workflow:

```bash
make workflow_test_release
```

To test all workflows:

```bash
make workflow_test_all
```

## Available Make Commands

For more information about each command, run `make help`.
