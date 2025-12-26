<template>
	<div class="title text-center">
		<div class="text-h6">
			{{ title }}
		</div>
		<p :style="{ color: '#5f708a' }">{{ desc }}</p>
	</div>
</template>

<script setup lang="ts">
import { defineComponent } from 'vue';

interface Props {
	title: string;
	desc: string;
}
withDefaults(defineProps<Props>(), {});
</script>

<style scoped></style>
