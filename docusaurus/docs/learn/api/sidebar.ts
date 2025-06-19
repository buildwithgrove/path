import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

const sidebar: SidebarsConfig = {
  apisidebar: [
    {
      type: "doc",
      id: "learn/api/path-path-api-toolkit-harness",
    },
    {
      type: "category",
      label: "API",
      items: [
        {
          type: "doc",
          id: "learn/api/1-handle-service-request",
          label: "EVM Service Request",
          className: "api-method post",
        },
        {
          type: "doc",
          id: "learn/api/2-health-check",
          label: "Health Check",
          className: "api-method get",
        },
        {
          type: "doc",
          id: "learn/api/3-disqualified-endpoints",
          label: "Disqualified Endpoints",
          className: "api-method get",
        },
      ],
    },
  ],
};

export default sidebar.apisidebar;
