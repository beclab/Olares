<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useData, withBase } from "vitepress";

const { frontmatter, lang, page } = useData();
const picker = ref<HTMLDetailsElement | null>(null);
const isZh = computed(() => lang.value.startsWith("zh"));
const current = computed(() => frontmatter.value.connectionVersion ?? "1.12.7");
const latestPath = computed(() => frontmatter.value.connectionLatestPath as string);
const label = computed(() => isZh.value ? "适用 Olares 版本" : "Olares version");
const options = computed(() => [
  { value: "1.12.7", label: "Olares 1.12.7+", href: withBase(latestPath.value) },
  { value: "1.12.6", label: "Olares 1.12.6", href: withBase(`${latestPath.value}-1.12.6`) },
]);
function closePicker() {
  if (picker.value) picker.value.open = false;
}

function onFocusOut(event: FocusEvent) {
  // A clicked link may not receive focus (for example, in Safari). Closing
  // on a null relatedTarget hides it before the click can navigate.
  if (event.relatedTarget instanceof Node && !picker.value?.contains(event.relatedTarget)) {
    closePicker();
  }
}

function onPointerDown(event: PointerEvent) {
  if (event.target instanceof Node && !picker.value?.contains(event.target)) {
    closePicker();
  }
}

onMounted(() => document.addEventListener("pointerdown", onPointerDown, true));
onBeforeUnmount(() => document.removeEventListener("pointerdown", onPointerDown, true));
watch(() => page.value.relativePath, closePicker);
</script>

<template>
  <nav v-if="latestPath" class="version-route-select" :aria-label="label">
    <details
      ref="picker"
      class="version-route-select__control"
      @keydown.esc.stop="closePicker"
      @focusout="onFocusOut"
    >
      <summary class="version-route-select__trigger">
        <span>{{ label }}: {{ current === '1.12.7' ? '1.12.7+' : current }}</span>
        <span aria-hidden="true">▾</span>
      </summary>
      <div class="version-route-select__menu">
        <a
          v-for="option in options"
          :key="option.value"
          class="version-route-select__option"
          :class="{ 'is-selected': option.value === current }"
          :href="option.href"
          :aria-current="option.value === current ? 'page' : undefined"
        >
          {{ option.label }}
        </a>
      </div>
    </details>
  </nav>
</template>

<style scoped>
.version-route-select { display: inline-flex; margin: 8px 0 12px; }
.version-route-select__control { position: relative; }
.version-route-select__trigger {
  display: inline-flex; align-items: center; gap: 7px; min-height: 32px;
  padding: 0 8px; border-radius: 7px; color: var(--vp-c-text-2);
  font-size: 13px; font-weight: 500; cursor: pointer; list-style: none;
}
.version-route-select__trigger::-webkit-details-marker { display: none; }
.version-route-select__trigger:hover,
details[open] .version-route-select__trigger { background: var(--vp-c-bg-soft); color: var(--vp-c-text-1); }
.version-route-select__trigger:focus-visible,
.version-route-select__option:focus-visible { outline: 2px solid var(--vp-c-brand-1); outline-offset: 2px; }
.version-route-select__menu {
  position: absolute; top: calc(100% + 6px); left: 0; z-index: 30;
  width: 208px; padding: 5px; background: var(--vp-c-bg-elv);
  border: 1px solid var(--vp-c-divider); border-radius: 10px; box-shadow: var(--vp-shadow-3);
}
.version-route-select__option {
  display: block; padding: 7px 9px; min-height: 34px; border-radius: 6px;
  color: var(--vp-c-text-1); font-size: 13px; line-height: 20px; text-decoration: none;
}
.version-route-select__option:hover { background: var(--vp-c-bg-soft); }
.version-route-select__option.is-selected { color: var(--vp-c-brand-1); background: var(--vp-c-brand-soft); font-weight: 600; }
</style>
