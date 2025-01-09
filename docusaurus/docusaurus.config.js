// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).
// There are various equivalent ways to declare your Docusaurus config.
// See: https://docusaurus.io/docs/api/docusaurus-config

import { themes as prismThemes } from "prism-react-renderer";

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Path",
  tagline: "All paths lead to Grove",
  favicon: "img/grove-leaf.jpeg",

  markdown: {
    mermaid: true,
  },
  themes: [
    "@docusaurus/theme-mermaid",
    [
      require.resolve("@easyops-cn/docusaurus-search-local"),
      /** @type {import('@easyops-cn/docusaurus-search-local').PluginOptions} **/
      {
        docsRouteBasePath: "/",
        hashed: false,
        indexBlog: false,
        highlightSearchTermsOnTargetPage: true,
        explicitSearchResultPath: true,
      },
    ],
  ],

  // Set the production url of your site here
  url: "https://grove.city",
  baseUrl: "/",

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: "buildwithgrove",
  projectName: "path",
  deploymentBranch: "gh-pages",
  trailingSlash: false,

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: "/",
          sidebarPath: "./sidebars.js",
          sidebarCollapsible: false,
        },
        theme: {
          customCss: "./src/css/custom.css",
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      docs: {
        sidebar: {
          hideable: false,
          autoCollapseCategories: false,
        },
      },
      style: "dark",
      navbar: {
        title: "Path",
        logo: {
          alt: "Path logo",
          src: "img/grove-leaf.jpeg",
        },
        items: [
          {
            type: "docSidebar",
            position: "left",
            sidebarId: "developSidebar",
            label: "💻 Develop",
          },
          {
            type: "docSidebar",
            position: "left",
            sidebarId: "operateSidebar",
            label: "⚙️ Operate",
          },
        ],
      },
      footer: {
        style: "dark",
        links: [
          {
            title: "Documentation",
            items: [
              {
                label: "Path",
                to: "/",
              },
              {
                label: "Path",
                href: "https://docs.grove.city/",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Discord - Grove",
                href: "https://discord.gg/build-with-grove",
              },
              {
                label: "Twitter",
                href: "https://twitter.com/buildwithgrove",
              },
            ],
          },
          {
            title: "More",
            items: [
              {
                label: "GitHub",
                href: "https://github.com/buildwithgrove/path",
              },
            ],
          },
        ],
        copyright: `Grove Inc.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
        additionalLanguages: [
          "gherkin",
          "protobuf",
          "json",
          "makefile",
          "diff",
          "lua",
          "bash",
        ],
      },
    }),
};

export default config;
