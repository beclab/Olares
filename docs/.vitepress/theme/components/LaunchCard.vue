<template>
  <div class="launch-card">
    <div class="card-content">
      <h2>
        {{ title }}
        <a :href="'#' + slugify(title)" aria-hidden="true" class="anchor-link">
          <svg xmlns="http://www.w3.org/2000/svg" fill="currentColor" width="24" height="24" viewBox="0 0 24 24">
            <path d="M0 0h24v24H0z" fill="none"></path>
            <path d="M3.9 12c0-1.71 1.39-3.1 3.1-3.1h4V7H7c-2.76 0-5 2.24-5 5s2.24 5 5 5h4v-1.9H7c-1.71 0-3.1-1.39-3.1-3.1zM8 13h8v-2H8v2zm9-6h-4v1.9h4c1.71 0 3.1 1.39 3.1 3.1s-1.39 3.1-3.1 3.1h-4V17h4c2.76 0 5-2.24 5-5s-2.24-5-5-5z"></path>
          </svg>
        </a>
      </h2>
      <p>{{ description }}</p>
      <ul>
        <li v-for="(link, index) in links" :key="index">
          <a :href="link.href">{{ link.text }}</a>
        </li>
      </ul>
      <button class="portal-button" :id="'btn-' + slugify(title)" @click="navigate">
        {{ buttonText }}
      </button>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    title: String,
    description: String,
    links: Array,
    buttonText: String,
    buttonLink: String
  },
  methods: {
    slugify(text) {
      return text.toLowerCase().replace(/\s+/g, '-');
    },
    navigate() {
      window.location.href = this.buttonLink;
    }
  }
}
</script>

<style>
.launch-card-container {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.launch-card {
  flex: 1 1 calc(33.333% - 1rem);
  margin-bottom: 1rem;
  display: flex;
  box-sizing: border-box;
}

.card-content {
  padding: 1rem;
}

.card-content h2 {
  border: none;
  margin-top: 0;
}

.anchor-link {
  visibility: hidden;
}

h2:hover .anchor-link {
  visibility: visible;
}

.portal-button {
  background-color: #ec6464;
  color: white;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 4px;
  cursor: pointer;
  margin-top: 1rem;
}

.portal-button:hover {
  background-color: #ec5a68;
}
</style>