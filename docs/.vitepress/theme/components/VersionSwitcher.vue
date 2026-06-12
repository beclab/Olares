<script setup lang="ts">
import { useData, useRouter,inBrowser } from "vitepress"
import { computed, ref } from 'vue'
import VPMenuLink from 'vitepress/dist/client/theme-default/components/VPMenuLink.vue'
import VPFlyout from 'vitepress/dist/client/theme-default/components/VPFlyout.vue'
 
const props = defineProps<{
  versions: string[]
  latestVersion: string
}>();

const router = useRouter();
const { site, page } = useData();

declare const __SITE_PATH_PREFIX__: string;
declare const __CURRENT_DOC_VERSION__: string;

function normalizePrefix(raw: string): string {
  if (!raw) return "/";
  let p = raw.startsWith("/") ? raw : `/${raw}`;
  if (!p.endsWith("/")) p += "/";
  return p;
}

function stripTrailingVersionSegment(base: string): string {
  let b = base || "/";
  if (!b.endsWith("/")) b += "/";
  for (const v of props.versions) {
    const segment = `${v}/`;
    if (b.endsWith(segment)) {
      return normalizePrefix(b.slice(0, b.length - segment.length));
    }
  }
  return normalizePrefix(b);
}

// Current pathname, made reactive to VitePress client-side navigation.
// Reading router.route.path registers a reactive dependency so dependent
// computeds recompute on every in-app navigation; window.location.pathname
// carries the full URL (base + version + page) and is already up to date by
// the time the route changes.
const currentPathname = computed(() => {
  void router.route.path;
  return inBrowser ? window.location.pathname : "/";
});

// Deploy prefix (e.g. /docs/). On versioned builds site.base is /docs/1.12.4/ —
// strip any version suffix; do not skip latestVersion (each branch build may set
// LATEST_VERSION to its own tag while base still carries that version segment).
const pathPrefix = computed(() => {
  if (typeof __SITE_PATH_PREFIX__ === "string" && __SITE_PATH_PREFIX__) {
    return normalizePrefix(__SITE_PATH_PREFIX__);
  }

  if (inBrowser) {
    const pathname = currentPathname.value;
    for (const v of props.versions) {
      const marker = `/${v}/`;
      const idx = pathname.indexOf(marker);
      if (idx > 0) {
        return normalizePrefix(pathname.slice(0, idx));
      }
    }
  }

  return stripTrailingVersionSegment(site.value.base || "/");
});

const originUrl = computed(() => {
  if (!inBrowser) return "";
  const url = window.location.origin;
  return url.endsWith("/") ? url : `${url}/`;
});

const currentVersion = computed(() => {
  if (
    typeof __CURRENT_DOC_VERSION__ === "string" &&
    __CURRENT_DOC_VERSION__ &&
    props.versions.includes(__CURRENT_DOC_VERSION__)
  ) {
    return __CURRENT_DOC_VERSION__;
  }

  if (inBrowser) {
    const pathname = currentPathname.value;
    for (const v of props.versions) {
      if (pathname.includes(`/${v}/`)) return v;
    }
  }

  const path = router.route.path;
  for (const v of props.versions) {
    if (path.startsWith(`/${v}/`)) return v;
  }
  return props.latestVersion;
});

// Path after the version segment (e.g. /manual/overview), for same-page jumps
// across versions. Derived from page.relativePath, which is reactive AND
// available during SSR prerender — so the built HTML keeps the current page
// instead of collapsing to each version root. relativePath is already relative
// to the docs (version) root, so there is no fragile base/version slicing.
const pathSuffixAfterVersion = computed(() => {
  const rel = page.value.relativePath || "";
  // Drop the source extension and map index pages to their directory.
  let p = rel.replace(/\.(md|html)$/i, "").replace(/(^|\/)index$/i, "$1");
  if (!p) return "/";
  return p.startsWith("/") ? p : `/${p}`;
});

const versionHref = (version: string) => {
  const prefix = pathPrefix.value;
  const base =
    version === props.latestVersion ? prefix : `${prefix}${version}/`;

  const suffix = pathSuffixAfterVersion.value;
  let path = base;
  if (suffix && suffix !== "/") {
    const tail = suffix.startsWith("/") ? suffix.slice(1) : suffix;
    path = base.endsWith("/") ? `${base}${tail}` : `${base}/${tail}`;
  }

  return inBrowser ? `${originUrl.value}${path.replace(/^\//, "")}` : path;
};

const isOpen = ref(false);
const toggle = () => {
  isOpen.value = !isOpen.value;
};
</script>

<template>
  <VPFlyout  class="VPVersionSwitcher" icon="vpi-versioning" :button="currentVersion"
    :label="'Switch Version'">
    <div class="items">
      <!-- <VPMenuLink v-if="currentVersion != latestVersion" :item="{
        text: latestVersion,
        link: `/`,
      }" /> -->
       <template v-for="version in versions" :key="version">
        <!-- <VPMenuLink v-if="currentVersion != version" :item="{
          text: version,
          link: `${localUrl}${version}/`,
          target: '_blank',
          rel: 'a'
        }" />   -->
       <a
         v-if="currentVersion != version"
         :href="versionHref(version)"
         target="_blank"
         rel="noopener noreferrer"
       >{{ version }}</a>
      </template>
    </div>
  </VPFlyout>
   
</template>

<style>
.vpi-versioning.option-icon {
  margin-right: 2px !important;
}

.vpi-versioning {
  --icon: url("data:image/svg+xml;charset=utf-8;base64,PHN2ZyB3aWR0aD0iNjRweCIgaGVpZ2h0PSI2NHB4IiB2aWV3Qm94PSIwIDAgMjQgMjQiIHN0cm9rZS13aWR0aD0iMi4yIiBmaWxsPSJub25lIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIGNvbG9yPSIjMDAwMDAwIj48cGF0aCBkPSJNMTcgN0MxOC4xMDQ2IDcgMTkgNi4xMDQ1NyAxOSA1QzE5IDMuODk1NDMgMTguMTA0NiAzIDE3IDNDMTUuODk1NCAzIDE1IDMuODk1NDMgMTUgNUMxNSA2LjEwNDU3IDE1Ljg5NTQgNyAxNyA3WiIgc3Ryb2tlPSIjMDAwMDAwIiBzdHJva2Utd2lkdGg9IjIuMiIgc3Ryb2tlLWxpbmVjYXA9InJvdW5kIiBzdHJva2UtbGluZWpvaW49InJvdW5kIj48L3BhdGg+PHBhdGggZD0iTTcgN0M4LjEwNDU3IDcgOSA2LjEwNDU3IDkgNUM5IDMuODk1NDMgOC4xMDQ1NyAzIDcgM0M1Ljg5NTQzIDMgNSAzLjg5NTQzIDUgNUM1IDYuMTA0NTcgNS44OTU0MyA3IDcgN1oiIHN0cm9rZT0iIzAwMDAwMCIgc3Ryb2tlLXdpZHRoPSIyLjIiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI+PC9wYXRoPjxwYXRoIGQ9Ik03IDIxQzguMTA0NTcgMjEgOSAyMC4xMDQ2IDkgMTlDOSAxNy44OTU0IDguMTA0NTcgMTcgNyAxN0M1Ljg5NTQzIDE3IDUgMTcuODk1NCA1IDE5QzUgMjAuMTA0NiA1Ljg5NTQzIDIxIDcgMjFaIiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS13aWR0aD0iMi4yIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjwvcGF0aD48cGF0aCBkPSJNNyA3VjE3IiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS13aWR0aD0iMi4yIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjwvcGF0aD48cGF0aCBkPSJNMTcgN1Y4QzE3IDEwLjUgMTUgMTEgMTUgMTFMOSAxM0M5IDEzIDcgMTMuNSA3IDE2VjE3IiBzdHJva2U9IiMwMDAwMDAiIHN0cm9rZS13aWR0aD0iMi4yIiBzdHJva2UtbGluZWNhcD0icm91bmQiIHN0cm9rZS1saW5lam9pbj0icm91bmQiPjwvcGF0aD48L3N2Zz4=")
}
</style>

<style scoped>
.VPVersionSwitcher {
  display: flex;
  align-items: center;
}



.icon {
  padding: 8px;
}

.title {
  padding: 0 24px 0 12px;
  line-height: 32px;
  font-size: 14px;
  font-weight: 700;
  color: var(--vp-c-text-1);
}




.VPScreenVersionSwitcher {
  border-bottom: 1px solid var(--vp-c-divider);
  height: 48px;
  overflow: hidden;
  transition: border-color 0.5s;
}

.VPVersionSwitcher a {
  display: block;
  border-radius: 6px;
  padding: 0 12px;
  line-height: 32px;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-text-1);
  white-space: nowrap;
  transition:
    background-color 0.25s,
    color 0.25s;
}

.VPVersionSwitcher a:hover {
  color: var(--vp-c-brand-1);
  background-color: var(--vp-c-default-soft);
}

.VPVersionSwitcher a.active {
  color: var(--vp-c-brand-1);
}


.VPScreenVersionSwitcher .items {
  visibility: hidden;
}

.VPScreenVersionSwitcher.open .items {
  visibility: visible;
}

.VPScreenVersionSwitcher.open {
  padding-bottom: 10px;
  height: auto;
}

.VPScreenVersionSwitcher.open .button {
  padding-bottom: 6px;
  color: var(--vp-c-brand-1);
}

.VPScreenVersionSwitcher.open .button-icon {
  /*rtl:ignore*/
  transform: rotate(45deg);
}

.VPScreenVersionSwitcher button .icon {
  margin-right: 8px;
}

.button {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 4px 11px 0;
  width: 100%;
  line-height: 24px;
  font-size: 14px;
  font-weight: 500;
  color: var(--vp-c-text-1);
  transition: color 0.25s;
}

.button:hover {
  color: var(--vp-c-brand-1);
}

.button-icon {
  transition: transform 0.25s;
}

.group:first-child {
  padding-top: 0px;
}

.group+.group,
.group+.item {
  padding-top: 4px;
}
</style>
