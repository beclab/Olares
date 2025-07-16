// docs/.vitepress/theme/index.ts
import DefaultTheme from "vitepress/theme";
import "./styles/custom.css";
import "./styles/index.css";
import { inBrowser, useRoute, useRouter, useData } from "vitepress";
import Layout from "./components/Layout.vue";
import { injectSpeedInsights } from "@vercel/speed-insights";
import { inject } from "@vercel/analytics";
import { App } from "vue";
import Tabs from "./components/tabs.vue";
import LaunchCard from "./components/LaunchCard.vue";
import FilterableList from "./components/FilterableList.vue";
import { onMounted, watch, nextTick, onBeforeMount,computed } from "vue";
import mediumZoom from "medium-zoom";
import OSTabs from "./components/OStabs.vue";
import VersionSwitcher from "./components/VersionSwitcher.vue";
import _ from "lodash";

const LANGUAGE_ZH_PATH = "/zh/";
const LANGUAGE_ZH_KEY = "zh";
const LANGUAGE_EN_KEY = "en";

const LANGUAGE_LOCAL_KEY = "language";
let isMenuChange = false;

export default {
  extends: DefaultTheme,
  Layout,
  enhanceApp({ app }: { app: App }) {
    app.component("Tabs", Tabs);
    app.component("LaunchCard", LaunchCard);
    app.component("FilterableList", FilterableList);
    app.component("OSTabs", OSTabs);
    app.component("VersionSwitcher", VersionSwitcher);
  },

  setup() {
    const route = useRoute();
    const router = useRouter();
    const { lang } = useData();

    const routerRedirect = () => {
      let localLanguage = localStorage.getItem(LANGUAGE_LOCAL_KEY) || 'en';
      
      const versions = process.env.VERSIONS!.split(",") ||[];
      versions.push('default');

      const languages = process.env.LANGUAGES!.split(",") || [];
      languages.push('en');
      console.log(versions, languages,localLanguage)

      if(!languages?.includes(localLanguage) ){
        localLanguage = 'en';
      }


      const currentPath = router.route.path;
      
      console.log('router.route.path', router.route.path);
      for( const l of languages ) {
        let localLanguagePath = (l === 'en' ? '' : `/${l}`);
        for (const v of versions) {
            let localVersionPath = (v === 'default' ? '' : `/${v}`);
            const u = `${localVersionPath}${localLanguagePath}`;
            console.log('checkPrefix', u);
            if (currentPath.startsWith(u)) {
                console.log('find localLanguage', localLanguage, l);
                if( l !== localLanguage ) {
                  let targetLanguagePath = (localLanguage === 'en' ? '' : `/${localLanguage}`);
                  const nextUrl = `${localVersionPath}${targetLanguagePath}${route.path.replace(u, '')}`;
                  router.go(nextUrl);
                }            
                return;
            
          }
        }
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
      inject();
      injectSpeedInsights();
      initZoom();

      document
        .querySelector(".wrapper .container a.title")
        ?.setAttribute("href", "https://www.olares.com/");

      document
        .querySelector(".wrapper .container a.title")
        ?.setAttribute("target", "_blank");
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

          document
            .querySelector(".wrapper .container a.title")
            ?.setAttribute("href", "https://www.olares.com/");

          document
            .querySelector(".wrapper .container a.title")
            ?.setAttribute("target", "_blank");
        });
      }
    );
  },
};
