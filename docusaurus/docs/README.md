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

Start by going through the [PATH Introduction](develop/path/introduction.md).
to learn how PATH works, how to configure it and how to run it locally.

## Is PATH for me?

If you're a Web2 Gateway Provider, you have four modes of operation to choose from:

1. **Dependent (Step Aside)**: Let clients use [Grove's Portal](https://portal.grove.city/) directly. For example, if you do not provide the services the customer is looking for.

2. **Grove Hybrid (Frontend)**: Provide a custom front-end experience and build your own gateway infrastructure but leverage [Grove's Portal API](https://docs.grove.city/) behind the **scenes**. For example, if you want want to build your own business logic around quality-of-service, load balancing, authentication as well as building your own front-end.

3. **PATH Hybrid (Full Stack):** Use `PATH` so you can provide the client with a customized end-to-end experience but also settle traffic on `Pocket Network` yourself without relying on `Grove`'s infrastructure at all.

4. **Independent (On Your Own):** Use your own stack to settle traffic on your own infrastructure, independent of `Grove`, `PATH` or `Pocket Network`.

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

    Mobile ---> | 1\. Dependent | Grove1
    Mobile ---> | 2\. Grove Hybrid| Grove2
    Mobile ---> | 3\. PATH Hybrid | Grove3
    Mobile ---> | 4\. Independent | Gateway

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

### Implementation Mode Comparison

If you're a Web2 Gateway Provider, you can use this table to understand you preferable mode of operation:

| Mode                         | Your Backend Infrastructure | Your Gateway Frontend | Your Gateway uses PATH | Customer uses Grove's Portal | Traffic Settled on Pocket Network | Description / Example                                                                                                         |
| ---------------------------- | --------------------------- | --------------------- | ---------------------- | ---------------------------- | --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| 1. Dependent (Step Aside)    | ❌                           | ❌                     | ❌                      | ✅                            | ✅                                 | Customers go to Grove's Portal for direct access                                                                              |
| 2. Grove Hybrid (Frontend)   | ❌                           | ✅                     | ❌                      | ❌                            | ✅                                 | Customers go to your frontend but use Grove's Portal API backend behind the scenes                                            |
| 3. PATH Hybrid (Full Stack)  | ❌                           | ✅                     | ✅                      | ❌                            | ✅                                 | Customers go to your frontend but use PATH's features (e.g. Quality-of Service) and settle traffic on Pocket network directly |
| 4. Independent (On Your Own) | ✅                           | ✅                     | ✅                      | ❌                            | ❌                                 | Customers go to your frontend and dependent on your gateway and infrastructure across the whole stack                         |
