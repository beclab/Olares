// docs/.vitepress/theme/index.ts
import DefaultTheme from "vitepress/theme";
import "./styles/custom.css";
import "./styles/index.css";
import { inBrowser, useRoute, useRouter, useData } from "vitepress";
import Layout from "./components/Layout.vue";
import { App } from "vue";
import Tabs from "./components/tabs.vue";
import LaunchCard from "./components/LaunchCard.vue";
import UseCaseGallery from "./components/UseCaseGallery.vue";
import { onMounted, watch, nextTick, onBeforeMount,computed } from "vue";
import mediumZoom from "medium-zoom";
import OSTabs from "./components/OStabs.vue";
import VersionSwitcher from "./components/VersionSwitcher.vue";
import _ from "lodash";
import { redirects, temporaryRedirects } from './redirects';

const clientRedirects: Record<string, string> = { ...redirects, ...temporaryRedirects };
import AppLinkGlobal from './components/AppLinkGlobal.vue'
import AppLinkCN from './components/AppLinkCN.vue'


const LANGUAGE_LOCAL_KEY = "language";
let isMenuChange = false;

export default {
  extends: DefaultTheme,
  Layout,
enhanceApp({ app, router }: { app: App; router: Router }) {
    app.component("Tabs", Tabs);
    app.component("LaunchCard", LaunchCard);
    app.component("UseCaseGallery", UseCaseGallery);
    app.component("OSTabs", OSTabs);
    app.component("VersionSwitcher", VersionSwitcher);
    app.component('AppLinkGlobal', AppLinkGlobal)
    app.component('AppLinkCN', AppLinkCN)

      router.onBeforeRouteChange = (to: string) => {
          const path = to.replace(/\.html$/i, ''),
              toPath = clientRedirects[path];

          if (toPath) {
              setTimeout(() => { router.go(toPath); })
              return false;
          } else {
              return true;
          }
      }
  },


  setup() {
    const route = useRoute();
    const router = useRouter();
    const { lang, site } = useData();

    // Auto-redirect a fresh page load to the visitor's remembered language.
    // router.route.path includes the site base (e.g. "/docs/" or
    // "/docs/1.12.4/"); the version, when present, lives entirely in that base.
    // We therefore strip the base first and match the language prefix on the
    // base-relative path. (The previous implementation matched against the raw
    // path without accounting for the base, so under a "/docs" deploy the
    // default+en combination produced an empty prefix that matched everything
    // and corrupted the URL, e.g. /docs/zh/... -> /zh/docs/zh/...).
    const routerRedirect = () => {
      let localLanguage = localStorage.getItem(LANGUAGE_LOCAL_KEY) || 'en';

      const languages = process.env.LANGUAGES ? process.env.LANGUAGES.split(",") : [];
      if (!languages.includes('en')) languages.push('en');

      if (!languages.includes(localLanguage)) {
        localLanguage = 'en';
      }

      const base = site.value.base || '/';
      const rawPath = router.route.path;
      // Base-relative path without a leading slash, e.g. "zh/manual/x" or
      // "manual/x". Handle the base both with and without its trailing slash:
      // router.route.path can be "/docs" as well as "/docs/".
      let rel: string;
      if (rawPath.startsWith(base)) {
        rel = rawPath.slice(base.length);
      } else if (base.endsWith('/') && rawPath === base.slice(0, -1)) {
        rel = '';
      } else {
        rel = rawPath.replace(/^\//, '');
      }

      // Detect the current language from the (non-en) prefix, if any.
      let currentLanguage = 'en';
      let pagePath = rel;
      for (const l of languages) {
        if (l === 'en') continue;
        if (rel === l || rel.startsWith(`${l}/`)) {
          currentLanguage = l;
          pagePath = rel.slice(l.length).replace(/^\//, '');
          break;
        }
      }

      if (currentLanguage === localLanguage) return;

      const langPrefix = localLanguage === 'en' ? '' : `${localLanguage}/`;
      const target = `${base}${langPrefix}${pagePath}`;
      if (target !== rawPath) {
        router.go(target);
      }
    };

    const initZoom = () => {
      mediumZoom(".main img", { background: "var(--vp-c-bg)" });
    };

    const toggleMenuStatus = () => {
      const menuDom = document.querySelector(".menu .VPMenu");
      menuDom?.addEventListener("click", (e) => {
        const target = e.target as Element;
        const isLink = target.closest(".VPMenuLink");
        if (isLink) {
          isMenuChange = true;
        }
      });
    };

    if (inBrowser) {
      routerRedirect();
    }

    onMounted(() => {
      toggleMenuStatus();
      initZoom();

      // document
      //   .querySelector(".wrapper .container a.title")
      //   ?.setAttribute("href", "https://www.olares.com/");

      // document
      //   .querySelector(".wrapper .container a.title")
      //   ?.setAttribute("target", "_blank");
    });

    watch(
      () => lang.value,
      (newValue) => {
        localStorage.setItem(LANGUAGE_LOCAL_KEY, newValue);
        isMenuChange = false;
      }
    );

    watch(
      () => route.path,
      () => {
        nextTick(() => {
          initZoom();

          // document
          //   .querySelector(".wrapper .container a.title")
          //   ?.setAttribute("href", "https://www.olares.com/");

          // document
          //   .querySelector(".wrapper .container a.title")
          //   ?.setAttribute("target", "_blank");
        });
      }
    );
  },
};