<script setup lang="ts">
const targetUrl = 'https://www.olares.cn/larepass'
</script>

<template>
  <a :href="targetUrl" target="_blank" rel="noreferrer" class="app-link">
    <slot>LarePass 下载页面</slot>
  </a>
</template>