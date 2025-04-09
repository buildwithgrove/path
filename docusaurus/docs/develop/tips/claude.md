---
sidebar_position: 2
title: Claude Sync
---

## Claude Sync <!-- omit in toc -->

This repository is set up to use [ClaudeSync](https://github.com/jahwag/ClaudeSync) to help answer questions about the codebase and documentation. Claude Sync enables developers to ask questions about the docs, streamlines customer support, makes documentation more discoverable, and allows for iterative improvements.

## Table of Contents <!-- omit in toc -->

- [Benefits](#benefits)
- [Getting Started](#getting-started)
  - [Installation](#installation)
  - [Authentication](#authentication)
  - [Creating a Project](#creating-a-project)
  - [Syncing Your Changes](#syncing-your-changes)
- [Available Commands](#available-commands)
- [Managing Categories](#managing-categories)
- [Ignoring Files](#ignoring-files)
- [System Prompt](#system-prompt)
- [Best Practices](#best-practices)

## Benefits

Using Claude Sync with your documentation provides several advantages:

- **Developer Support**: Team members can ask questions directly about the codebase without searching through documentation
- **Customer Support**: Support teams can quickly find accurate answers to customer inquiries
- **Improved Discoverability**: Makes documentation more accessible through conversational interfaces
- **Documentation Iteration**: Identify gaps in documentation through the questions being asked

## Getting Started

### Installation

Ensure you have Python set up on your machine, then install ClaudeSync:

```shell
pip install claudesync
```

### Authentication

Follow the instructions in your terminal to authenticate:

```shell
claudesync auth login
```

### Creating a Project

Initialize a new ClaudeSync project using:

```shell
make claudesync_init
```

This command will:

1. Check if ClaudeSync is installed
2. Guide you through creating a new project
3. Provide instructions for setting up the system prompt

### Syncing Your Changes

After making changes to your documentation, sync them with Claude:

```shell
make claudesync_push
```

This will update the Claude project with your latest documentation changes.

## Available Commands

The following Make targets are available:

- `make claudesync_init`: Initialize a new ClaudeSync project
- `make claudesync_push`: Push all changes to your Claude project
- `make claudesync_categories`: List all available categories
- `make claudesync_add_category`: Add a new category for specific file types
- `make claudesync_push_category`: Push only files from a specific category

## Managing Categories

Categories allow you to organize your files for more targeted syncing:

```shell
# List existing categories
make claudesync_categories

# Add a new category
make claudesync_add_category

# Push only specific categories
make claudesync_push_category
```

## Ignoring Files

The `.claudeignore` file controls which files are excluded from syncing. This ensures Claude's context is limited to relevant documentation.

Common patterns to exclude:

- Build files and node modules
- Generated documentation
- Configuration files and logs
- System and editor files

## System Prompt

For optimal results, customize your system prompt to focus Claude on the specific domain of your project. A well-crafted system prompt should:

1. Define Claude's specialty area
2. Specify the type of assistance required
3. Provide formatting guidelines for responses
4. Set technical focus areas
5. List topics to avoid

Here's a template to start with:

```text
You are a documentation assistant specialized in technical documentation.
Your primary role is to provide clear explanations about the project's functionality, architecture, and usage patterns based on the documentation you have access to.

When answering questions:
- Always reference specific documentation sections
- Provide code examples when relevant
- Highlight best practices and recommended approaches
- Link to related documentation pages
- Provide step-by-step guides for complex procedures

Present your analysis and recommendations in this format:
- Begin with a concise summary of the answer
- List key points using bullet points when appropriate
- Provide code examples with explanatory comments when needed
- Include references to specific documentation files
- Conclude with next steps or related topics to explore

Technical guidance should focus on:
- Installation and setup procedures
- Configuration options and their impact
- Common usage patterns and workflows
- Troubleshooting common issues
- Integration points with other systems
- Best practices for using the project

Avoid:
- Speculating beyond what's in the documentation
- Providing outdated information
- Giving opinions not supported by the documentation
- Discussing implementation details not covered in the docs

Remember that the user may be new to the project or an experienced developer. Adjust your explanations based on the complexity of their questions while maintaining technical accuracy and completeness.
```

## Best Practices

1. **Regular Updates**: Sync after significant documentation changes
2. **Structured Documentation**: Well-organized docs make Claude more effective
3. **Specific Categories**: Create categories for different documentation types
4. **Test Questions**: Ask sample questions to verify Claude's understanding
5. **Iterative Improvement**: Use Claude's responses to identify documentation gaps
