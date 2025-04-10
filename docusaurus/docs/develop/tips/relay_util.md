---
sidebar_position: 1
title: Relay Util
description: Simple Load Testing Tool
---

:::tip

Easy to use make targets are provided in the test requests [Makefile](https://github.com/buildwithgrove/path/blob/main/makefiles/test_requests.mk).

`make test_request__shannon_relay_util_1000`

`make test_request__shannon_relay_util_10000`

These targets send the given number of requests to `localhost:3069` for the `anvil` service on Shannon.

**For additional configuration options, see the docs below.**

:::

import RemoteMarkdown from '@site/src/components/RemoteMarkdown';

<RemoteMarkdown src="https://raw.githubusercontent.com/commoddity/relay-util/refs/heads/main/README.md" />
