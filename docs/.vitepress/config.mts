import { defineConfig,UserConfig,DefaultTheme } from "vitepress";
import { withMermaid } from "vitepress-plugin-mermaid";
import { en } from "./en";
import { zh } from "./zh";
import _ from "lodash";
import { redirects, temporaryRedirects } from "./theme/redirects";
//import defaultConfig from 'vitepress-versioning-plugin';

// Paths collected during transformPageData so sitemap.transformItems can
// filter them out without re-reading frontmatter from disk.
const noindexPaths = new Set<string>();

// Normalize a route/path to the extensionless, slash-trimmed form used for
// sitemap comparisons (works regardless of cleanUrls or trailing slashes).
const normalizeRoute = (p: string) =>
  p.replace(/^\/+/, "").replace(/\.html$/, "").replace(/\/$/, "");

// Every path that 301/302-redirects at the edge (see theme/redirects.ts) must
// be kept out of sitemap.xml: a sitemap URL that redirects is flagged by Google
// as "Page with redirect" and wastes crawl budget.
const redirectSources = new Set<string>(
  [...Object.keys(redirects), ...Object.keys(temporaryRedirects)].map(normalizeRoute)
);

// Deep, low-value reference pages that inflate "Discovered – currently not
// indexed" without real search demand. noindex them in both locales so crawl
// budget concentrates on the pages users actually search for. Matched against
// the locale-stripped route (so it covers /docs/... and /docs/zh/... alike).
const noindexRoutePattern =
  /^developer\/contribute\/olares-id\/(contract|verifiable-credential)\//;

// Archived (non-latest) versioned doc builds deploy under /docs/<version>/
// (e.g. /docs/1.12.4/...) and duplicate the latest docs almost verbatim,
// which is the main driver of the "duplicate content" / "missing canonical"
// findings. Mark every page in such a build `noindex` so the versioned copies
// drop out of search and consolidate onto the canonical latest /docs/... URLs.
// Detected from the deploy base, the only signal reliably set for archived
// builds: it always carries a version segment (e.g. "/docs/1.12.4/") and can
// never match the latest deploy, whose base is "/docs/" or "/" (no version).
const isArchivedVersionBuild = /\/\d+\.\d+/.test(process.env.BASE_URL || "");

 

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

  // With a single published version, every non-latest copy 301s to /docs/, so
  // the switcher would render an empty dropdown (its only entry equals the
  // current version and is filtered out). Skip injecting it entirely.
  if( !versions || versions.length <= 1 ) {
    return config;
  }

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
  titleTemplate: ":title | Olares Docs",
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
    logoLink: { link: "https://www.olares.com/", target: "_self" },
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
    // Archived versioned builds noindex every page (see isArchivedVersionBuild).
    const localeStrippedRoute = pageData.relativePath.replace(/^zh\//, "");
    const noindexByPath = noindexRoutePattern.test(localeStrippedRoute);
    if (isArchivedVersionBuild || pageData.frontmatter?.noindex || noindexByPath) {
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
      // Skip canonical on noindex pages: a self/duplicate canonical combined
      // with `noindex` sends contradictory signals to crawlers.
      return;
    }

    // Self-referencing canonical for every indexable doc page. Built from the
    // *unversioned* route on the production origin (matches the sitemap
    // hostname below) so versioned (/docs/<version>/...) duplicates consolidate
    // onto the latest /docs/... URL. `cleanUrls` is on, so the route is
    // extensionless; directory indexes keep their trailing slash.
    const route = pageData.relativePath
      .replace(/(^|\/)index\.md$/, '$1')
      .replace(/\.md$/, '');
    const canonicalHref = route
      ? `https://www.olares.com/docs/${route}`
      : 'https://www.olares.com/docs/';
    pageData.frontmatter.head ??= [];
    pageData.frontmatter.head.push([
      'link',
      { rel: 'canonical', href: canonicalHref },
    ]);
    // Keep og:url identical to the canonical URL. A mismatch between the two
    // sends conflicting signals to crawlers and social scrapers about the
    // page's authoritative address.
    const isZh = pageData.relativePath.startsWith('zh/');
    const siteName = isZh ? 'Olares 文档' : 'Olares Docs';
    const defaultDescription = isZh
      ? '让人们重新拥有自己的数据'
      : 'Let people own their data again';
    const ogTitle = pageData.title
      ? `${pageData.title} | ${siteName}`
      : siteName;
    const ogDescription =
      pageData.description?.trim() || defaultDescription;
    // Same brand OG asset as marketing (from Olares One / Shopify, 1200×630).
    const ogImage = 'https://www.olares.com/docs/olares-og.jpg';
    const ogLocale = isZh ? 'zh_CN' : 'en_US';
    for (const [property, content] of [
      ['og:url', canonicalHref],
      ['og:title', ogTitle],
      ['og:description', ogDescription],
      ['og:image', ogImage],
      ['og:image:width', '1200'],
      ['og:image:height', '630'],
      ['og:image:type', 'image/jpeg'],
      ['og:type', 'website'],
      ['og:site_name', siteName],
      ['og:locale', ogLocale],
    ] as const) {
      pageData.frontmatter.head.push([
        'meta',
        { property, content },
      ]);
    }
    for (const [name, content] of [
      ['twitter:card', 'summary_large_image'],
      ['twitter:title', ogTitle],
      ['twitter:description', ogDescription],
      ['twitter:image', ogImage],
    ] as const) {
      pageData.frontmatter.head.push([
        'meta',
        { name, content },
      ]);
    }
  },

  // Archived versioned builds noindex every page, so their sitemap would be
  // empty. The `sitemap` package throws `EmptySitemap` when the stream ends
  // with zero URLs, which crashes the whole build (exit != 0) and makes the
  // release-docs pipeline silently skip that version. Since we explicitly do
  // not want archived versions in the sitemap anyway, disable it for them.
  sitemap: isArchivedVersionBuild ? undefined : {
    hostname: "https://www.olares.com/docs/",
    transformItems: (items) =>
      // Drop noindex pages from sitemap.xml so crawlers don't even discover
      // them via the sitemap. The meta tag above is what ultimately removes
      // them from search engine indexes; this just avoids the extra hit.
      items.filter((item) => {
        // Normalize to the extensionless route so the comparison holds whether
        // or not cleanUrls is enabled.
        const p = item.url.replace(/^\/+/, '').replace(/\.html$/, '');
        // Repo READMEs are srcExclude'd above, but guard the sitemap too so a
        // stray /docs/README (or nested README) can never leak back in.
        if (p === 'README' || p.endsWith('/README')) return false;
        // Drop URLs that 301/302-redirect at the edge — a sitemap entry that
        // redirects is what Google reports as "Page with redirect".
        if (redirectSources.has(p.replace(/\/$/, ''))) return false;
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
