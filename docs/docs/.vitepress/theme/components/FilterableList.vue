<template>
  <div class="filterable-list">
    <!-- Tag filters -->
    <div class="filters">
      <label
        v-for="(tag, index) in uniqueTags"
        :key="index"
        class="filter-label"
      >
        <input
          type="checkbox"
          :value="tag"
          v-model="selectedTags"
        />
        {{ tag }}
      </label>
    </div>

    <!-- List of items -->
    <ul>
      <li v-for="(item, index) in filteredItems" :key="index">
        <a
          :href="item.link"
          :target="isExternalLink(item.link) ? '_blank' : '_self'"
          :rel="isExternalLink(item.link) ? 'noopener noreferrer' : ''"
        >
          {{ item.title }}
        </a>
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  name: 'FilterableList',
  props: {
    items: {
      type: Array,
      required: true,
      // Example structure of items:
      // [
      //   { title: 'Guide 1', link: '/guide1.md', tags: ['tag1', 'tag2'] },
      //   { title: 'Guide 2', link: '/guide2.md', tags: ['tag2', 'tag3'] },
      //   ...
      // ]
    }
  },
  data() {
    return {
      selectedTags: [] // Stores the currently selected tags for filtering
    }
  },
  computed: {
    // Compute the unique tags from the items
    uniqueTags() {
      const allTags = this.items.flatMap(item => item.tags)
      return [...new Set(allTags)]
    },
    // Filtered list based on selected tags
    filteredItems() {
      if (this.selectedTags.length === 0) {
        return this.items // No filtering if no tags are selected
      }
      return this.items.filter(item =>
        item.tags.some(tag => this.selectedTags.includes(tag))
      )
    }
  },
  methods: {
    // Check if the link is external by looking for 'http' or 'https'
    isExternalLink(link) {
      return /^(http|https):\/\//.test(link)
    }
  }
}
</script>

<style scoped>
/* Simple styling for the filterable list */
.filterable-list {
  padding: 20px;
}

.filters {
  margin-bottom: 20px;
}

.filter-label {
  margin-right: 10px;
  cursor: pointer;
}

ul {
  list-style: none;
  padding: 0;
}

ul li {
  margin: 10px 0;
}

ul li a {
  text-decoration: none;
  color: var(--vp-c-link);
}

</style>