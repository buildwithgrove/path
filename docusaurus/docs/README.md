<div align="center">
<h1>PATH<br/>Path API & Toolkit Harness</h1>
<img src="https://storage.googleapis.com/grove-brand-assets/Presskit/Logo%20Joined-2.png" alt="Grove logo" width="500"/>

</div>
<br/>

:::warning

🚧 This documentation is still under construction 🚧

:::

## Overview

**PATH** (Path API & Toolkit Harness) is an open source framework for enabling
access to a decentralized supply network.

It provides various tools and libraries to streamline the integration and
interaction with decentralized protocols.

## Getting Started

### PATH

Start by going through the [PATH Walkthrough](develop/path/introduction.md).
to learn how PATH works, how to configure it and how to run it locally.

### Envoy

Start by going through the [Envoy Walkthrough](./develop/envoy/walkthrough.md).
to learn how our Envoy integration works, how to configure it and how to run it locally.

## Is PATH for me?

If you're a Web2 Gateway Provider, you have four modes of operation to choose from:

1. **Step aside (N/A)**: Let clients use [Grove's Portal](https://portal.grove.city/) directly. For example, if you do not provide the services the client is looking for.

2. **Hybrid (Frontline)**: Provide a front-end using your custom stack but leverage [Grove's Portal API](https://docs.grove.city/) behind the scenes. For example, if you want to provide a custom front-end but leverage Grove's infrastructure.

3. **Full Stack (Direct):** Use `PATH` so you can provide the client with a custom experience but also settle traffic on `Pocket Network` yourself without relying on Grove's infrastructure.

4. **Independent (N/A):** Use your own stack to settle traffic on your own infrastructure, independent of `Grove`, `PATH` or `Pocket Network`.

Here's the information reorganized into a table and nodes section:

```mermaid
%%{init: {'flowchart': {'curve': 'basis', 'lineWidth': 2}, 'themeVariables': {'fontFamily': 'arial', 'nodeTextColor': '#000000', 'edgeTextColor': '#000000'}}}%%

flowchart LR
    Mobile((📱))

    subgraph Routes
        Grove1["To Pocket Network <br> Through Grove 🌿 <br> Using PATH"]
        Grove2["To Pocket Network <br> Through Web2 Gateway <br> Via Grove 🌿"]
        Grove3["To Pocket Network <br> Through Web2 Gateway <br> Using PATH"]
        Gateway["To Web2 Gateway's Servers <br> Through Web2 Gateway <br> Using Gateway's Stack"]
    end

    subgraph PubSrv["Public Servers"]
        Server1[(Public Server 1)]
        Server2[(Public Server 2)]
    end

    subgraph PrivSrv["Private Servers"]
        PServer1[(Private Server 1)]
        PServer2[(Private Server 2)]
    end

    PubLabel["Pocket Network <br> - Shared <br> - Permissionless <br> - Incentivized <br> - Open Source Providers"]
    PrivLabel["Web2 Gateway Servers <br> -Private <br> - Dedicated <br> - Solely owned"]

    Mobile ---> | 1 | Grove1
    Mobile ---> | 2 | Grove2
    Mobile ---> | 3 | Grove3
    Mobile ---> | 4 | Gateway

    Grove1 --> Server1
    Grove1 --> Server2

    Grove2 -.-> Grove1

    Grove3 ---> Server1
    Grove3 ---> Server2

    Gateway ---> PServer1
    Gateway ---> PServer2


    %% Label connections
    PubSrv -.- PubLabel
    PrivSrv -.- PrivLabel

    %% Styling
    classDef publicServers fill:#3D9970,stroke:#333,stroke-width:2px,color:#000000
    classDef privateServers fill:#E67E22,stroke:#333,stroke-width:2px,color:#000000
    classDef groveRoute fill:#E8F5E9,stroke:#333,stroke-width:2px,color:#000000
    classDef web2Route fill:#FFEBEE,stroke:#333,stroke-width:2px,color:#000000
    classDef labelStyle fill:none,stroke:none
    class Server1,Server2 publicServers
    class PServer1,PServer2 privateServers
    class Grove1,Grove2 groveRoute
    class Grove3,Gateway web2Route
    class PubLabel,PrivLabel labelStyle
```

### Implementation Modes

| Mode                   | Your Infrastructure | Your Gateway | Your Frontend | PATH | Grove | Pocket Network | Description                                                     |
| ---------------------- | ------------------- | ------------ | ------------- | ---- | ----- | -------------- | --------------------------------------------------------------- |
| 1. Step Aside (N/A)    | ❌                  | ❌           | ❌            | ❌   | ✅    | ✅             | Provider directs clients to Grove's Portal for direct access    |
| 2. Hybrid (Frontline)  | ❌                  | ❌           | ✅            | ❌   | ✅    | ✅             | Provider uses custom frontend with Grove's Portal API backend   |
| 3. Full Stack (Direct) | ✅                  | ✅           | ✅            | ✅   | ❌    | ✅             | Provider uses PATH to settle traffic directly on Pocket Network |
| 4. Independent (N/A)   | ✅                  | ✅           | ✅            | ❌   | ❌    | ❌             | Provider uses entirely their own infrastructure and stack       |
