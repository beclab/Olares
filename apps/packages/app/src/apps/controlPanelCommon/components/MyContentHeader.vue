<template>
	<ListItem icon-size="md" class="q-px-lg my-content-header-wrapper">
		<template #avatar>
			<q-icon v-if="icon" size="20px" :name="icon" color="ink-1" />
			<q-img
				v-if="img"
				:src="img"
				:ratio="1"
				width="20px"
				style="width: 20px; height: 20px; border-radius: 50%"
			/>
		</template>
		<template #title>
			<slot v-if="$slots.title" name="title"></slot>
			<span v-else class="text-h6">{{ title }}</span>
		</template>
	</ListItem>
</template>

<script setup lang="ts">
import ListItem from './MyListItem/ListItem.vue';
import { useRoute } from 'vue-router';

interface Props {
	title?: string;
	icon?: string;
	img?: string;
}

defineProps<Props>();
const route = useRoute();

console.log('route', route.params);
</script>

<style lang="scss" scoped>
.my-content-header-wrapper {
	height: $content-header-height;
	.title {
		font-size: 14px;
	}
}
</style>
