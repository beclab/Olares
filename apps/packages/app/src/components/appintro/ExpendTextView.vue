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
		{{ showDescription ? lessText : resolvedMoreText }}
	</span>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, watch } from 'vue';
import { encode } from 'he';
import { useI18n } from 'vue-i18n';

type MarkdownIt = {
	render: (src: string) => string;
	utils?: {
		escapeHtml?: (src: string) => string;
	};
};

const props = defineProps({
	text: {
		type: String,
		default: '',
		require: true
	},
	markdown: {
		type: Boolean,
		default: false
	},
	highlight: {
		type: Boolean,
		default: false
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
		default: '',
		require: false
	},
	lessText: {
		type: String,
		default: '',
		require: false
	}
});

const { t } = useI18n();

const showDescription = ref(false);
const needMore = ref(false);

const markdownRenderer = ref<MarkdownIt | null>(null);
const markdownRendererMode = ref<'plain' | 'highlight' | null>(null);

async function ensureMarkdownRenderer() {
	// @ts-ignore - bundler supports dynamic import for code-splitting
	const [{ default: markdownit }] = await Promise.all([import('markdown-it')]);
	const desiredMode = props.highlight ? 'highlight' : 'plain';
	if (markdownRenderer.value && markdownRendererMode.value === desiredMode) {
		return;
	}

	if (!props.highlight) {
		markdownRenderer.value = markdownit({
			html: false,
			linkify: true,
			typographer: true,
			breaks: true
		});
		markdownRendererMode.value = 'plain';
		return;
	}

	// @ts-ignore - bundler supports dynamic import for code-splitting
	const [{ default: hljs }] = await Promise.all([import('highlight.js')]);
	// NOTE: highlight css is intentionally not imported here to avoid pulling
	// extra style dependencies into pages that don't need it.
	// If you want highlighted code blocks to look nicer, import a theme css
	// at the page/app level when `highlight` is enabled.

	const md: any = markdownit({
		html: false,
		linkify: true,
		typographer: true,
		breaks: true,
		highlight: (str: string, lang: string) => {
			try {
				if (lang && typeof (hljs as any).getLanguage === 'function') {
					const ok = (hljs as any).getLanguage(lang);
					if (ok) {
						return (
							'<pre><code class="hljs">' +
							(hljs as any).highlight(str, {
								language: lang,
								ignoreIllegals: true
							}).value +
							'</code></pre>'
						);
					}
				}
			} catch (e) {
				console.error(e);
			}

			const escaped =
				typeof md.utils?.escapeHtml === 'function'
					? md.utils.escapeHtml(str)
					: encode(str);
			return '<pre><code class="hljs">' + escaped + '</code></pre>';
		}
	});
	markdownRenderer.value = md;
	markdownRendererMode.value = 'highlight';
}

const formattedDescription = computed(() => {
	const raw = String(props.text ?? '');
	if (!props.markdown) {
		return encode(raw).replace(/\r\n|\n|\r/g, '<br/>');
	}

	const renderer = markdownRenderer.value;
	if (!renderer) {
		// Fallback while renderer is still loading.
		return encode(raw).replace(/\r\n|\n|\r/g, '<br/>');
	}

	return renderer.render(raw);
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

const resolvedMoreText = computed(() => props.moreText || t('base.more'));

onMounted(() => {
	checkTextOverflow();
});

watch(
	() => [props.markdown, props.highlight, props.text] as const,
	([markdownEnabled, _highlightEnabled, text]) => {
		if (markdownEnabled && String(text ?? '').trim().length > 0) {
			ensureMarkdownRenderer()
				.then(() => {
					checkTextOverflow();
				})
				.catch((e) => {
					console.error(e);
				});
		}
	},
	{ immediate: true }
);

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
