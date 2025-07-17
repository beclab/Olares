<template>
  <div class="os-tabs">
    <button
      v-for="tab in tabs"
      :key="tab.id"
      :class="{ active: currentTab === tab.id }"
      @click="handleTabClick(tab)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>

<script>
export default {
  props: {
    defaultTab: {
      type: String,
      default: "linux", // Default tab if no match is found
    },
    tabs: {
      type: Array,
      required: true,
      // Expected format:
      // [
      //   { id: "linux", label: "Linux", href: "/docs/linux" },
      //   { id: "macos", label: "Mac", href: "/docs/macos" },
      //   { id: "windows", label: "Windows", href: "/docs/windows" }
      // ]
    },
  },
  data() {
    return {
      currentTab: this.defaultTab, // Active tab starts with the defaultTab
    };
  },
  mounted() {
    this.setCurrentTabBasedOnURL(); // Set the active tab based on the current URL
  },
  methods: {
    handleTabClick(tab) {
      if (tab.href) {
        // Redirect to the tab's URL (full page reload)
        window.location.assign(tab.href);
      } else {
        // Otherwise, just switch the active tab locally
        this.currentTab = tab.id;
      }
    },
    setCurrentTabBasedOnURL() {
      // Get the current URL path
      const currentPath = window.location.pathname;

      // Match the current path with the `href` of the tabs
      const matchingTab = this.tabs.find((tab) => tab.href === currentPath);

      if (matchingTab) {
        this.currentTab = matchingTab.id; // Set the matching tab as active
      } else {
        // Fallback to the default tab if no match is found
        this.currentTab = this.defaultTab;
      }
    },
  },
};
</script>

<style>
.os-tabs {
  display: flex;
  margin-top: 1rem;
  margin-bottom: 1rem;
}

.os-tabs button {
  flex: 1;
  padding: 0.5rem 1rem;
  background-color: var(--vp-c-bg);
  color: var(--vp-c-text-primary);
  border: none;
  cursor: pointer;
  text-align: center;
  transition: background-color 0.3s, color 0.3s, border-bottom 0.3s;
  border-bottom: 2px solid transparent;
}

.os-tabs button.active {
  background-color: var(--vp-c-surface);
  color: var(--vp-c-brand);
  border-bottom-color: var(--vp-c-brand);
}

.os-tabs button:hover {
  background-color: var(--vp-c-surface);
  color: var(--vp-c-brand);
}
</style>
