<!-- docs/.vitepress/theme/Layout.vue -->
<script setup lang="ts">
import DefaultTheme from "vitepress/theme";
import { useData } from "vitepress";
import { watchEffect } from "vue";

const { lang } = useData();
watchEffect(() => {
  if (typeof window !== "undefined") {
    document.cookie = `nf_lang=${lang.value}; expires=Mon, 1 Jan 2030 00:00:00 UTC; path=/`;
  }
});
</script>

<template>
  <DefaultTheme.Layout />
</template>

<style>
.medium-zoom-overlay,
.medium-zoom-image {
  z-index: 9999 !important;
}
</style>
