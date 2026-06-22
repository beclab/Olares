<template>
  <div class="uc-gallery">
    <!-- Keyword (category) filter pills -->
    <div class="uc-filters">
      <button
        v-for="cat in filters"
        :key="cat"
        type="button"
        :class="['uc-pill', { 'is-active': selected === cat }]"
        @click="selected = cat"
      >
        {{ catLabel(cat) }}
      </button>
    </div>

    <!-- Free-text search -->
    <input
      v-model="query"
      type="search"
      class="uc-search"
      :placeholder="ui.searchPlaceholder"
      :aria-label="ui.searchPlaceholder"
    />

    <!-- Filtered results -->
    <div v-if="filtered.length" class="uc-cards">
      <a
        v-for="item in filtered"
        :key="item.link"
        class="uc-card"
        :href="localizeLink(item.link)"
      >
        <span class="uc-card-head">
          <span class="uc-card-title">{{ displayTitle(item) }}</span>
          <span class="uc-card-cat">{{ catLabel(item.category) }}</span>
        </span>
        <span class="uc-card-desc">{{ displayDesc(item) }}</span>
      </a>
    </div>
    <p v-else class="uc-empty">{{ ui.empty }}</p>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { withBase, useData } from "vitepress";
import {
  useCases,
  useCaseCategories,
  categoryLabelsZh,
} from "../../data/useCases";

const { lang } = useData();
const isZh = computed(() => lang.value.toLowerCase().startsWith("zh"));

const filters = ["All", ...useCaseCategories];
const selected = ref("All");
const query = ref("");

const ui = computed(() => {
  return isZh.value
    ? {
        searchPlaceholder: "搜索应用示例…",
        empty: "没有匹配的应用示例",
      }
    : {
        searchPlaceholder: "Search use cases…",
        empty: "No use cases match your search",
      };
});

function catLabel(cat) {
  if (cat === "All") {
    return isZh.value ? "全部" : "All";
  }
  return isZh.value ? categoryLabelsZh[cat] ?? cat : cat;
}

function displayTitle(item) {
  return isZh.value && item.titleZh ? item.titleZh : item.title;
}

function displayDesc(item) {
  return isZh.value && item.descriptionZh
    ? item.descriptionZh
    : item.description;
}

function localizeLink(link) {
  const prefix = isZh.value ? "/zh" : "";
  return withBase(prefix + link);
}

const filtered = computed(() => {
  const q = query.value.trim().toLowerCase();
  return useCases.filter((item) => {
    const matchesCategory =
      selected.value === "All" || item.category === selected.value;
    const title = displayTitle(item).toLowerCase();
    const desc = displayDesc(item).toLowerCase();
    const matchesQuery = !q || title.includes(q) || desc.includes(q);
    return matchesCategory && matchesQuery;
  });
});
</script>

<style scoped>
.uc-gallery {
  margin: 1.5rem 0 3rem;
}

.uc-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.uc-pill {
  padding: 0.3rem 0.85rem;
  font-size: 0.85rem;
  line-height: 1.4;
  border: 1px solid var(--vp-c-divider);
  border-radius: 999px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-2);
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease, background-color 0.2s ease;
}

.uc-pill:hover {
  border-color: var(--vp-c-text-3);
  color: var(--vp-c-text-1);
}

.uc-pill.is-active {
  background-color: var(--vp-c-text-1);
  border-color: var(--vp-c-text-1);
  color: var(--vp-c-bg);
}

.uc-search {
  width: 100%;
  max-width: 360px;
  padding: 0.5rem 0.85rem;
  margin-bottom: 1.5rem;
  font-size: 0.9rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 8px;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-1);
  outline: none;
  transition: border-color 0.2s ease;
}

.uc-search:focus {
  border-color: var(--vp-c-text-3);
}

.uc-cards {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 1rem;
}

.uc-card {
  display: flex;
  flex-direction: column;
  padding: 1.1rem 1.25rem;
  border: 1px solid var(--vp-c-divider);
  border-radius: 12px;
  background-color: var(--vp-c-bg);
  text-decoration: none;
  transition: box-shadow 0.25s ease;
}

.uc-card:hover {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.08);
}

.uc-card-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 0.5rem;
  margin-bottom: 0.4rem;
}

.uc-card-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--vp-c-text-1);
}

.uc-card-cat {
  flex-shrink: 0;
  font-size: 0.72rem;
  color: var(--vp-c-text-3);
  white-space: nowrap;
}

.uc-card-desc {
  font-size: 0.88rem;
  line-height: 1.5;
  color: var(--vp-c-text-2);
}

.uc-empty {
  color: var(--vp-c-text-3);
}

.dark .uc-card:hover {
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.3);
}
</style>
