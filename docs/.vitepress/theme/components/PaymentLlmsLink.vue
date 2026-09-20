<template>
  <p class="payment-llms-link">
    <a :href="href">{{ href }}</a>
  </p>
</template>

<script setup lang="ts">
import { withBase } from "vitepress";
import { computed, onMounted, ref } from "vue";

const origin = ref("");
onMounted(() => {
  origin.value = location.origin;
});

const href = computed(() => {
  const path = withBase("/developer/payment/llms-full.txt");
  return origin.value ? `${origin.value}${path}` : path;
});
</script>

<style scoped>
.payment-llms-link {
  margin: 0.5rem 0 1rem;
  word-break: break-all;
}
.payment-llms-link a {
  font-family: var(--vp-font-family-mono);
  font-size: 0.9em;
}
</style>
