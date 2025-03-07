---
sidebar_position: 2
title: Helm Chart Values
description: Chart values for the GUARD Helm chart
---

The following values may be used to customize the GUARD Helm chart.

For example, you may run the following command to install the GUARD Helm chart with a custom value for the `authServer.enabled` key:

```bash
helm install guard buildwithgrove/guard --set authServer.enabled=false
```

Or you may use the following command to upgrade the GUARD Helm chart with a custom `values.yaml` file:

```bash
helm upgrade guard buildwithgrove/guard -f values.yaml
```

import RemoteMarkdown from '@site/src/components/RemoteMarkdown';

<!-- TODO_IMPROVE(@commoddity): Update this to point to main branc once PR # 20 merged -->
<RemoteMarkdown src="https://raw.githubusercontent.com/buildwithgrove/helm-charts/refs/heads/guard-helm-charts/charts/guard/README.md" />
