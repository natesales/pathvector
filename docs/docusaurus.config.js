const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').DocusaurusConfig} */
module.exports = {
  title: 'Pathvector',
  tagline: 'Edge Routing Platform',
  url: 'https://pathvector.io',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/icon-white.png',
  organizationName: 'natesales', // Usually your GitHub org/user name.
  projectName: 'pathvector', // Usually your repo name.
  themeConfig: {
    navbar: {
      title: 'Pathvector',
      // logo: {
      //   alt: 'Pathvector Logo',
      //   src: 'img/icon-white.svg',
      // },
      items: [
        {
          type: 'doc',
          docId: 'about',
          position: 'left',
          label: 'Docs',
        },
        // {to: '/blog', label: 'Blog', position: 'left'},
        {
          href: 'https://github.com/natesales/pathvector',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      links: [
        // {
        //   title: 'Community',
        //   items: [
        //     {
        //       label: 'GitHub',
        //       href: 'https://github.com/natesales/pathvector',
        //     },
        //   ],
        // },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Nate Sales.`,
    },
    prism: {
      theme: lightCodeTheme,
      darkTheme: darkCodeTheme,
    },
  },
  presets: [
    [
      '@docusaurus/preset-classic',
      {
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl:
            'https://github.com/facebook/docusaurus/edit/master/website/',
        },
        blog: {
          showReadingTime: true,
          // Please change this to your repo.
          editUrl:
            'https://github.com/facebook/docusaurus/edit/master/website/blog/',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      },
    ],
  ],
};
