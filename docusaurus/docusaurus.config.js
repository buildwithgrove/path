// @ts-check
// `@type` JSDoc annotations allow editor autocompletion and type checking
// (when paired with `@ts-check`).
// There are various equivalent ways to declare your Docusaurus config.
// See: https://docusaurus.io/docs/api/docusaurus-config

import { themes as prismThemes } from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Path',
  tagline: 'All paths lead to Grove',
  favicon: 'img/grove-leaf.jpeg',

  // Set the production url of your site here
  url: 'https://grove.city',
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'buildwithgrove',
  projectName: 'path',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          routeBasePath: '/',
          sidebarPath: './sidebars.js',
          sidebarCollapsible: false,
        },
        theme: {
          customCss: './src/css/custom.css',
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
      style: 'dark',
      navbar: {
        title: 'Path',
        logo: {
          alt: 'Path logo',
          src: 'img/grove-leaf.jpeg',
        },
        // TODO_UPNEXT: Add documentation sidebars about operation, contributing etc.
        items: [],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Documentation',
            items: [
              {
                label: 'Path',
                to: '/',
              },
              {
                label: 'Path',
                href: 'https://docs.grove.city/',
              },
            ],
          },
          {
            title: 'Community',
            items: [
              {
                label: 'Discord - Grove',
                href: 'https://discord.gg/uRnKAufk',
              },
              {
                label: 'Twitter',
                href: 'https://twitter.com/buildwithgrove',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/buildwithgrove/path',
              },
            ],
          },
        ],
        copyright: `Grove Inc.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
      },
    }),
};

export default config;
