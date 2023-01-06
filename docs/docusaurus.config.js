// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = {
  plain: {
      color: "#393A34",
      backgroundColor: "#f6f8fa"
  },
  styles: [{
      types: ["comment", "prolog", "doctype", "cdata"],
      style: {
          color: "#999988",
          fontStyle: "italic"
      }
  }, {
      types: ["namespace"],
      style: {
          opacity: 0.7
      }
  }, {
      types: ["string", "attr-value"],
      style: {
          color: "#bc00ff"
      }
  }, {
      types: ["punctuation", "operator"],
      style: {
          color: "#393A34"
      }
  }, {
      types: ["entity", "url", "symbol", "number", "boolean", "variable", "constant", "property", "regex", "inserted"],
      style: {
          color: "#aa38ff"
      }
  }, {
      types: ["atrule", "keyword", "attr-name", "selector"],
      style: {
          color: "#505050"
      }
  }, {
      types: ["function", "deleted", "tag"],
      style: {
          color: "#d73a49"
      }
  }, {
      types: ["function-variable"],
      style: {
          color: "#6f42c1"
      }
  }, {
      types: ["tag", "selector", "keyword"],
      style: {
          color: "#00009f"
      }
  }]
};

let darkCodeTheme = {
  plain: {
      color: "#F8F8F2",
      backgroundColor: "#191919"
  },
  styles: [{
      types: ["prolog", "constant", "builtin"],
      style: {
          color: "rgb(189, 147, 249)"
      }
  }, {
      types: ["inserted", "function"],
      style: {
          color: "rgb(80, 250, 123)"
      }
  }, {
      types: ["deleted"],
      style: {
          color: "rgb(255, 85, 85)"
      }
  }, {
      types: ["changed"],
      style: {
          color: "rgb(255, 184, 108)"
      }
  }, {
      types: ["punctuation", "symbol"],
      style: {
          color: "rgb(248, 248, 242)"
      }
  }, {
      types: ["string", "char", "tag", "selector"],
      style: {
          color: "#bc00ff"
      }
  }, {
      types: ["keyword", "variable"],
      style: {
          color: "rgb(189, 147, 249)",
          fontStyle: "italic"
      }
  }, {
      types: ["comment"],
      style: {
          color: "rgb(98, 114, 164)"
      }
  }, {
      types: ["attr-name"],
      style: {
          color: "rgb(241, 250, 140)"
      }
  }]
};

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'Pathvector',
  tagline: 'Edge Routing Platform',
  url: 'https://pathvector.io',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'img/icon-white.png',
  organizationName: 'natesales', // Usually your GitHub org/user name.
  projectName: 'pathvector', // Usually your repo name.

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
          sidebarPath: require.resolve('./sidebars.js'),
          // Remove this to remove the "edit this page" links.
          editUrl: 'https://github.com/natesales/pathvector/edit/main/docs/',
        },
        blog: false,
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      announcementBar: {
        id: 'announcement',
        content: 'Has Pathvector helped automate your network? Please consider <a href="https://github.com/sponsors/natesales">supporting the project</a>. A small donation goes a long way to keep Pathvector sustainable and free for everyone.',
        backgroundColor: '#dd00ff',
        textColor: '#2d2d2d',
        isCloseable: false,
      },
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
            {
                type: 'doc',
                docId: 'installation',
                position: 'left',
                label: 'Install',
            },
            {
                type: 'doc',
                docId: 'configuration',
                position: 'left',
                label: 'Configuration',
            },
            {
                href: 'https://github.com/natesales/pathvector',
                label: 'GitHub',
                position: 'right',
            },
        ],
    },
      footer: {
        links: [],
        copyright: `Copyright © ${new Date().getFullYear()} Nate Sales.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
    }),
};

module.exports = config;
