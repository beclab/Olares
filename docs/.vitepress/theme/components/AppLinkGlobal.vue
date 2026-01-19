<script setup lang="ts">
const targetUrl = 'https://www.olares.com/larepass'
</script>

<template>
  <a :href="targetUrl" target="_blank" rel="noreferrer" class="app-link">
    <slot>LarePass Download page</slot>
  </a>
</template>