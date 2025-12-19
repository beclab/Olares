<template>
	<div
		class="text-body2 text-ink-2"
		:class="showDescription ? 'multi-row' : 'break-all'"
		:style="{ '--displayLine': displayLine, '--expendAlign': align }"
		ref="multiRow"
		v-html="formattedDescription"
	/>
	<span
		v-if="needMore"
		class="text-body2 text-info cursor-pointer"
		@click="toggleDescription"
	>
		{{ showDescription ? lessText : moreText }}
	</span>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { i18n } from '@apps/market/src/boot/i18n';
import { encode } from 'he';

const props = defineProps({
	text: {
		type: String,
		default: '',
		require: true
	},
	displayLine: {
		type: Number,
		default: 3
	},
	align: {
		type: String,
		default: 'left'
	},
	moreText: {
		type: String,
		default: i18n.global.t('base.more'),
		require: false
	},
	lessText: {
		type: String,
		default: '',
		require: false
	}
});

const showDescription = ref(false);
const needMore = ref(false);

const formattedDescription = computed(() => {
	return encode(props.text).replace(/\r\n|\n|\r/g, '<br/>');
});

const multiRow = ref<HTMLElement | null>(null);

const checkTextOverflow = () => {
	if (multiRow.value) {
		const height = multiRow.value.scrollHeight;
		const lineHeight = parseInt(getComputedStyle(multiRow.value).lineHeight);

		needMore.value = height > lineHeight * props.displayLine;
	}
};

const toggleDescription = () => {
	showDescription.value = !showDescription.value;
	checkTextOverflow();
};

onMounted(() => {
	checkTextOverflow();
});

watch(formattedDescription, () => {
	checkTextOverflow();
});
</script>

<style scoped lang="scss">
.multi-row {
	max-width: 100%;
	width: 100%;
	overflow: visible;
	text-align: var(--expendAlign);
}

.break-all {
	max-width: 100%;
	width: 100%;
	text-align: var(--expendAlign);
	display: -webkit-box;
	overflow: hidden;
	white-space: normal !important;
	text-overflow: ellipsis;
	/* autoprefixer: off */
	word-wrap: break-word;
	-webkit-line-clamp: var(--displayLine);
	-webkit-box-orient: vertical;
}
</style>
