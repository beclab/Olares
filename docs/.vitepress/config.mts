import { defineConfig,UserConfig,DefaultTheme } from "vitepress";
import { withMermaid } from "vitepress-plugin-mermaid";
import { en } from "./en";
import { zh } from "./zh";
import _ from "lodash";
//import defaultConfig from 'vitepress-versioning-plugin';

// Paths collected during transformPageData so sitemap.transformItems can
// filter them out without re-reading frontmatter from disk.
const noindexPaths = new Set<string>();

 

function defineVersionedConfig2(
  defaultConfig: UserConfig<DefaultTheme.Config>
): UserConfig<DefaultTheme.Config> {
  let config = _.defaultsDeep(defaultConfig);

  if( !process.env.BASE_URL || !process.env.VERSIONS || !process.env.LATEST_VERSION ) {
    return config;
  }

  const versions =  process.env.VERSIONS?.split(",");
  const latestVersion = process.env.LATEST_VERSION || versions![versions!.length - 1];
  console.log(versions, latestVersion);

  for( const locale of Object.keys(config.locales) ) {
    let themeConfig = config.locales[locale]!.themeConfig!;

    themeConfig?.nav?.push(     
          {
            component: 'VersionSwitcher',
            // Optional props to pass to the component
            props: {
              versions,
              latestVersion,
            }
          }
    );
  }
  
  return config;
}



// https://vitepress.dev/reference/site-config
export default defineVersionedConfig2(withMermaid({
  title: "Olares",
  description: "Let people own their data again",
  lang: "en",
  locales: {
    root: {
      label: "English",
      ...en,
    },
    zh: {
      label: "简体中文",
      ...zh,
    },
  },
  themeConfig: {
    logoLink: "https://www.olares.com/",
    editLink: {
      pattern: "https://github.com/beclab/Olares/edit/main/docs/:path",
      text: "Edit this page on GitHub",
    },
    search: {
      provider: "algolia",
      options: {
        appId: "DZ6H2FVQGO",
        apiKey: "e5257d88b605dc0e5b82b12854aea9a5",
        indexName: "olares",
        searchParameters: {
          queryLanguages: ["zh", "en"],
          facetFilters: ['version:' + (process.env.CURRENT_VERSION || "main")]
        },
        locales: {
          zh: {
            placeholder: "搜索文档",
            translations: {
              button: {
                buttonText: "搜索文档",
                buttonAriaLabel: "搜索文档",
              },
              modal: {
                searchBox: {
                  resetButtonTitle: "清除查询条件",
                  resetButtonAriaLabel: "清除查询条件",
                  cancelButtonText: "取消",
                  cancelButtonAriaLabel: "取消",
                },
                startScreen: {
                  recentSearchesTitle: "搜索历史",
                  noRecentSearchesText: "没有搜索历史",
                  saveRecentSearchButtonTitle: "保存至搜索历史",
                  removeRecentSearchButtonTitle: "从搜索历史中移除",
                  favoriteSearchesTitle: "收藏",
                  removeFavoriteSearchButtonTitle: "从收藏中移除",
                },
                errorScreen: {
                  titleText: "无法获取结果",
                  helpText: "你可能需要检查你的网络连接",
                },
                footer: {
                  selectText: "选择",
                  navigateText: "切换",
                  closeText: "关闭",
                  searchByText: "搜索提供者",
                },
                noResultsScreen: {
                  noResultsText: "无法找到相关结果",
                  suggestedQueryText: "你可以尝试查询",
                  reportMissingResultsText: "你认为该查询应该有结果？",
                  reportMissingResultsLinkText: "点击反馈",
                },
              },
            },
          },
        },
      },
    },
  },

  transformPageData(pageData) {
    // Opt a page out of Google/Bing/Algolia indexing by adding `noindex: true`
    // to its frontmatter. Implemented here (rather than per-file `head`) so
    // we don't shift source line numbers, which would break `@include` ranges
    // in files that embed this one as a snippet.
    if (pageData.frontmatter?.noindex) {
      pageData.frontmatter.head ??= [];
      pageData.frontmatter.head.push([
        'meta',
        { name: 'robots', content: 'noindex, nofollow' },
      ]);
      // Store the extensionless route so this matches the sitemap item.url
      // regardless of `cleanUrls` (which drops the .html from item.url).
      noindexPaths.add(
        pageData.relativePath.replace(/(^|\/)index\.md$/, '$1').replace(/\.md$/, '')
      );
    }
  },

  sitemap: {
    hostname: "https://www.olares.com/docs/",
    transformItems: (items) =>
      // Drop noindex pages from sitemap.xml so crawlers don't even discover
      // them via the sitemap. The meta tag above is what ultimately removes
      // them from search engine indexes; this just avoids the extra hit.
      items.filter((item) => {
        // Normalize to the extensionless route so the comparison holds whether
        // or not cleanUrls is enabled.
        const p = item.url.replace(/^\/+/, '').replace(/\.html$/, '');
        return !noindexPaths.has(p);
      }),
  },
  lastUpdated: true,
  cleanUrls: true,
  // Snippet-only fragments and repo READMEs are pulled into real pages via
  // `<!--@include-->` (which reads the raw file directly, so excluding them as
  // routes does NOT break includes). Keeping them out of the build stops them
  // from leaking into routes, search, and sitemap.xml as junk URLs.
  srcExclude: [
    "**/README.md",
    "**/reusables/**",
    "**/reusables.md",
    "**/reusables-*.md",
    "**/*.reusables.md",
  ],
  base: process.env.BASE_URL || "/",
  vite: {
    build: {
      minify: "terser",
      chunkSizeWarningLimit: Infinity,
    },
    define: {
      'process.env.VERSIONS': JSON.stringify(process.env.VERSIONS || JSON.stringify([])),
      'process.env.LANGUAGES': JSON.stringify(process.env.LANGUAGES || JSON.stringify([])),
      // Deploy path prefix without version (e.g. /docs). Versioned builds set
      // BASE_URL=/docs/1.12.4/ so site.base alone cannot yield /docs/ for links.
      __SITE_PATH_PREFIX__: JSON.stringify(process.env.SITE_PATH_PREFIX || ''),
      __CURRENT_DOC_VERSION__: JSON.stringify(process.env.CURRENT_VERSION || ''),
    }
  },
  head: [
    [
      "link",
      {
        rel: "icon",
        href: (process.env.BASE_URL || "/") + "icon1.png",
      },
    ],
    [
      "script",
      {
        async: "",
        src: "https://www.googletagmanager.com/gtag/js?id=G-GMCVWRBP24",
      },
    ],
    [
      "script",
      {},
      `window.dataLayer = window.dataLayer || [];
      function gtag(){dataLayer.push(arguments);}
      gtag('js', new Date());
      gtag('config', 'G-GMCVWRBP24');`,
    ],
    [
      "meta",
      {
         name:"docsearch:version",
         content: process.env.CURRENT_VERSION || "main"
      },
    ],
    // 引入 Material Design Icons
    [
      "link",
      {
        rel: "stylesheet",
        href: "https://fonts.googleapis.com/css2?family=Material+Symbols+Outlined:opsz,wght,FILL,GRAD@20..48,200,0..1,-50..200",
      },
    ],
  ],
}));
