<template>
  <div class="tabs-container">
    <slot></slot>
    <div class="tabs">
      <button
        v-for="(tab, index) in tabLabels"
        :key="tab"
        @click="clickHandler(tab, index)"
        :class="{ active: activeTab === tab }"
      >
        <div class="tabs-item-wrapper">
          <img
            :src="iconFilter(index)"
            class="tabs-img-wrapper"
            alt=""
            v-if="iconFilter(index)"
          />
          <span> {{ tab }}</span>
        </div>
      </button>
    </div>
    <div v-for="tab in tabLabels" :key="tab" v-show="activeTab === tab">
      <div v-if="activeTab === tab" class="tab-item">
        <slot :name="tabSlots[tab]"></slot>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    isDark: {
      type: Boolean,
    },
    icons: {
      type: Array,
      default: [],
    },
  },
  data() {
    return {
      activeTab: null,
      tabLabels: [],
      tabSlots: {},
      randomKey: Math.random(),
    };
  },
  methods: {
    clickHandler(tab, index) {
      this.activeTab = tab;
      this.$emit("tab-changed", tab, index);
    },
    iconFilter(index) {
      if (this.icons[index]) {
        return this.isDark
          ? `/images/manual/icons/${this.icons[index]}-dark.svg`
          : `/images/manual/icons/${this.icons[index]}.svg`;
      }
    },
  },
  mounted() {
    // Map slot names to display labels
    const slots = Object.keys(this.$slots).filter((slot) => slot !== "default");
    this.tabSlots = slots.reduce((map, slot) => {
      const label = slot.replace(/-/g, " "); // Replace hyphens with spaces
      map[label] = slot;
      return map;
    }, {});
    console.log("aa", this.tabSlots);

    this.tabLabels = Object.keys(this.tabSlots);
    this.activeTab = this.tabLabels[0];
  },
};
</script>

<style>
.tabs {
  display: flex;
  background-color: var(--vp-c-bg);
}

.tabs button {
  flex: 1;
  padding: 0.5rem 1rem;
  border: none;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-primary);
  cursor: pointer;
  transition: background-color 0.3s, color 0.3s;
  border-bottom: 2px solid transparent;
}

.tabs button.active {
  background-color: var(--vp-c-surface);
  color: var(--vp-c-brand);
  border-bottom-color: var(--vp-c-brand);
}

.tabs button:hover {
  background-color: var(--vp-c-surface);
  color: var(--vp-c-brand);
}

div[style] {
  color: var(--vp-c-text-primary);
}
.tabs-img-wrapper {
  width: 12px;
  pointer-events: none;
}
.tabs-item-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;
}
.tabs-container .tab-item {
  margin-top: 16px;
}
</style>
