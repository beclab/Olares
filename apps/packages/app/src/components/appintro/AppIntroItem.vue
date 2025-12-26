<template>
	<div
		v-if="
			content ||
			link ||
			(linkArray && linkArray.length > 0) ||
			(contentArray && contentArray.length > 0)
		"
		class="info-item-root row justify-between items-start no-wrap"
	>
		<div class="info-title col-5 text-body2 text-ink-3">{{ title }}</div>
		<expend-text-view v-if="content" :text="content" align="right" />
		<div v-else-if="link" class="info_link col-7" @click="emit('onLinkClick')">
			{{ link }}
		</div>
		<div
			v-else-if="contentArray"
			class="info-content text-body2 text-ink-2 col-7"
		>
			{{ contentArray.join(separator) }}
		</div>
		<div v-else-if="linkArray" class="column justify-end col-7">
			<div
				v-for="(item, index) in linkArray"
				:key="index"
				@click="openUrl(item.url)"
				class="info_link"
				style="max-width: 100%"
			>
				{{ item.text }}
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onMounted, PropType, ref } from 'vue';
import ExpendTextView from '@apps/market/src/components/appintro/ExpendTextView.vue';

const props = defineProps({
	title: {
		type: String,
		required: true
	},
	content: {
		type: String,
		required: false
	},
	link: {
		type: String,
		required: false
	},
	separator: {
		type: String,
		default: ',',
		required: false
	},
	contentArray: {
		type: Object as PropType<string[]>,
		required: false
	},
	linkArray: Object as PropType<
		{
			text: string;
			url: string;
		}[]
	>
});

const expendDesc = ref();

const emit = defineEmits(['onLinkClick']);

onMounted(() => {
	if (props.content) {
		expendDesc.value = props.content;
	}
});

const openUrl = (url: string) => {
	if (url) {
		window.open(url);
	}
};
</script>

<style scoped lang="scss">
.info-item-root {
	width: 100%;
	height: auto;

	.info-title {
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	.info-content {
		text-align: right;
		overflow: hidden;
		text-overflow: ellipsis;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
	}

	.info_link {
		@extend .info-content;
		color: $info;
		cursor: pointer;
		text-decoration: none;
	}
}
</style>
