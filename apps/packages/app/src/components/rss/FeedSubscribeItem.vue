<template>
	<div
		class="feed-item-root row justify-start items-center q-py-xs"
		:class="subscribed ? '' : 'cursor-pointer'"
		@click="onClick"
	>
		<feed-icon size="32px" :feed="feed" />
		<div class="feed-info column full-height justify-between q-mx-sm">
			<span class="text-ellipsis text-subtitle3 text-ink-2">{{
				feed.title
			}}</span>
			<span class="text-ellipsis text-overline text-ink-3">{{
				feed.feed_url
			}}</span>
		</div>
		<q-icon
			v-if="subscribed"
			color="orange-default"
			size="20px"
			name="sym_r_bookmark_added"
			class="q-ml-xs"
		/>
		<bt-check-box
			v-else
			:model-value="modelValue"
			@update:model-value="(value) => emit('update:modelValue', value)"
			:circle="true"
		/>
	</div>
</template>

<script lang="ts" setup>
import FeedIcon from './FeedIcon';
import { PropType } from 'vue';
import { SearchFeed } from '../../utils/rss-types';
import BtCheckBox from '../rss/BtCheckBox.vue';
const props = defineProps({
	modelValue: {
		type: Boolean,
		required: true
	},
	subscribed: {
		type: Boolean,
		default: false
	},
	feed: {
		type: Object as PropType<SearchFeed>
	}
});

const emit = defineEmits(['update:modelValue']);

const onClick = () => {
	if (props.subscribed) {
		return;
	}
	emit('update:modelValue', !props.modelValue);
};
</script>

<style lang="scss" scoped>
.feed-item-root {
	height: 40px;
	width: 100%;

	.feed-info {
		overflow: hidden;
		width: calc(100% - 16px - 32px - 32px);
		max-width: calc(100% - 16px - 32px - 32px);

		.text-ellipsis {
			overflow: hidden;
			text-overflow: ellipsis;
			display: -webkit-box;
			-webkit-line-clamp: 1;
			-webkit-box-orient: vertical;
		}
	}
}
</style>
