---
sidebar_position: 2
title: Troubleshooting & FAQ
description: Tips & Tricks for Troubleshooting with FAQ
---

## Table of Contents <!-- omit in toc -->

- [How can I profile (`pprof`) PATH's runtime?](#how-can-i-profile-pprof-paths-runtime)

## How can I profile (`pprof`) PATH's runtime?

Use the `debug_goroutines` make target to view go runtime's info on PATH like so:

```bash
make debug_goroutines
```

You can then view the Golang's runtime at [localhost:8081/ui](http://localhost:8081/ui).
