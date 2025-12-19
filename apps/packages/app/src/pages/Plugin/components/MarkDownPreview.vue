<template>
	<div class="my-md-preview-wrapper-container text-body2">
		<div :style="{ minHeight: loadingHidden ? 0 : '16px' }">
			<div
				v-html="html"
				ref="popRef"
				:style="{ marginBottom: !loading && html ? '-20px' : 0 }"
			></div>
			<q-spinner-dots
				v-show="loading"
				color="grey-5"
				size="xs"
				style="line-height: 16px; height: 16px"
			/>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { computed, ref, toRefs } from 'vue';
import hljs from 'highlight.js';
import MarkdownIt from 'markdown-it';
import 'highlight.js/styles/a11y-light.min.css';
import { nextTick } from 'process';
import copyIcon from 'assets/images/content_copy.svg';

const md: MarkdownIt = MarkdownIt({
	highlight: function (str: string, lang: string) {
		if (lang && hljs.getLanguage(lang)) {
			try {
				return `<div class="hl-code"><div class="hl-code-header"><span>${lang}</span><div class="copy-container"><img src=${copyIcon} ><span class="code-copy">copy</span></div></div><div class="hljs"><code>${
					hljs.highlight(str, {
						language: lang,
						ignoreIllegals: true
					}).value
				}</code></div></div>`;
			} catch (__) {
				console.error(__, 'error');
			}
		}
		return `<div class="hl-code"><div class="hl-code-header"><span>${lang}</span></div><div class="hljs"><code>${md.utils.escapeHtml(
			str
		)}</code></div></div>`;
	}
});

interface Props {
	text: string;
	loading?: boolean;
	ended?: boolean;
	loadingHidden?: boolean;
}
const props = withDefaults(defineProps<Props>(), {
	loading: false,
	ended: true,
	loadingHidden: false
});

const { loading, loadingHidden } = toRefs(props);
const popRef = ref();

const html = computed(() => {
	if (!props.loading && !props.ended) {
		nextTick(() => {
			// TODO: code debounce
			// appendInputCursor();
		});
	}

	return md.render(props.text);
});
</script>
<style lang="scss" scoped>
.my-md-preview-wrapper-container {
	display: inline-block;
	// border-radius: 20px;
	border-top-left-radius: 20px;
	border-bottom-left-radius: 4px;
	border-top-right-radius: 20px;
	border-bottom-right-radius: 20px;
	color: $grey-10;

	::v-deep(.hl-code) {
		border-radius: 12px;
		overflow: hidden;
	}
	::v-deep(.hl-code-header) {
		padding: 8px 12px;
		background-color: $grey-2;
		display: flex;
		align-items: center;
		justify-content: space-between;
	}
	::v-deep(.copy-container) {
		display: flex;
		align-items: center;
		justify-content: space-between;
		cursor: pointer;
		color: $grey-8;
	}
	::v-deep(.code-copy) {
		margin-left: 4px;
	}
	::v-deep(.hljs) {
		padding: 12px;
		overflow-y: auto;
	}
	::v-deep(p) {
		margin-bottom: 20px;
	}
	::v-deep(pre) {
		margin: 20px 0;
	}
	::v-deep(code) {
		background: white;
		border-radius: 4px;
		margin: 0 4px;
	}
	::v-deep(code:only-child) {
		display: block;
		width: 286px;
		padding: 0;
		margin: 0;
		border-radius: 0px;
		overflow-x: scroll;
		background: transparent;
	}
}
</style>
<style lang="scss">
@keyframes blink {
	0% {
		opacity: 1;
	}
	50% {
		opacity: 0;
	}
	100% {
		opacity: 1;
	}
}

.input-cursor {
	animation: blink 1s infinite;
	position: relative;
	&::after {
		content: '';
		position: absolute;
		top: 4px;
		width: 4px;
		height: 14px;
		background: $grey-8;
		border-radius: 1px;
	}
}
</style>
