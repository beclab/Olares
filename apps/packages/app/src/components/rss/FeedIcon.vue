<template>
	<q-img
		:style="{ '--size': size }"
		class="feed-icon"
		:src="
			icon
				? icon
				: getFeedIcon(feed)
				? getFeedIcon(feed)
				: getRequireImage('feed_default_icon.svg')
		"
	>
		<template v-slot:loading>
			<q-skeleton :height="size" :width="size" type="circle" />
		</template>
		<template v-slot:error>
			<img class="feed-icon" :src="getRequireImage('feed_default_icon.svg')" />
		</template>
	</q-img>
</template>

<script lang="ts" setup>
import { Feed, SearchFeed } from 'src/utils/rss-types';
import { getRequireImage, getFeedIcon } from 'src/utils/rss-utils';
import { PropType } from 'vue';

defineProps({
	feed: {
		type: Object as PropType<Feed | SearchFeed>
	},
	icon: {
		type: String
	},
	size: {
		type: String,
		require: '20px'
	}
});
</script>

<style scoped lang="scss">
.feed-icon {
	width: var(--size) !important;
	height: var(--size) !important;
}
</style>
