---
sidebar_position: 1
title: PATH Configurations
description: Introduction to PATH Configurations
---

A `PATH` stack is configured via two files:

| File           | Required | Description                                                | LocalNet Location           |
| -------------- | -------- | ---------------------------------------------------------- | --------------------------- |
| `.config.yaml` | ✅       | PATH **Gateway & Service** configurations                  | `./local/path/.config.yaml` |
| `.values.yaml` | ❌       | PATH **Request Authorization & Deployment** configurations | `./local/path/.values.yaml` |
