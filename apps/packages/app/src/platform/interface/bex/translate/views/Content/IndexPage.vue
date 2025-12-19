<template>
	<span v-if="loading">
		{{ gap }}
		<LoadingIcon />
	</span>

	<span v-else-if="text && !sameLang">
		<template v-if="keeps.length > 0">
			{{ gap }}
			<span v-html="processedText" v-bind="styles" />
		</template>
		<template v-else>
			{{ gap }}
			<span v-bind="styles">{{ text }}</span>
		</template>
	</span>
</template>

<script setup>
import { computed, onMounted, onUnmounted, watch } from 'vue';
import LoadingIcon from './LoadingIcon.vue';
import {
	OPT_STYLE_LINE,
	OPT_STYLE_DOTLINE,
	OPT_STYLE_DASHLINE,
	OPT_STYLE_WAVYLINE,
	OPT_STYLE_FUZZY,
	OPT_STYLE_HIGHLIGHT,
	OPT_STYLE_BLOCKQUOTE,
	OPT_STYLE_DIY,
	DEFAULT_COLOR,
	MSG_TRANS_CURRULE,
	APP_LCNAME
} from '../../config';
import { useTranslate } from '../../hooks/Translate';
import interpreter from '../../libs/interpreter';

const props = defineProps({
	q: String,
	keeps: {
		type: Array,
		default: () => []
	},
	translator: Object,
	// eslint-disable-next-line vue/no-reserved-keys
	element: Object
});

const LINE_STYLES = {
	[OPT_STYLE_LINE]: 'solid',
	[OPT_STYLE_DOTLINE]: 'dotted',
	[OPT_STYLE_DASHLINE]: 'dashed',
	[OPT_STYLE_WAVYLINE]: 'wavy'
};

const { text, sameLang, loading, rule } = useTranslate(
	props.q,
	props.translator.rule,
	props.translator.setting
);

const handleKissEvent = (e) => {
	const { action, args } = e.detail;
	if (action === MSG_TRANS_CURRULE) {
		rule.value = args;
	}
};

onMounted(() => {
	window.addEventListener(props.translator.eventName, handleKissEvent);
});

onUnmounted(() => {
	window.removeEventListener(props.translator.eventName, handleKissEvent);
});

watch([() => text.value, () => rule.value.transEndHook], () => {
	if (text.value && rule.value.transEndHook?.trim()) {
		interpreter.run(`exports.transEndHook = ${rule.value.transEndHook}`);
		interpreter.exports.transEndHook(
			props.element,
			props.q,
			text.value,
			props.keeps
		);
	}
});

const gap = computed(() => {
	if (rule.value.transOnly === 'true') {
		return '';
	}
	return props.q.length >= props.translator.setting.newlineLength ? '\n' : ' ';
});

const styles = computed(() => ({
	'data-style': rule.value.textStyle,
	style: getStyleByType(rule.value.textStyle)
}));

const processedText = computed(() => {
	return text.value.replace(/\[(\d+)\]/g, (_, p) => props.keeps[parseInt(p)]);
});

watch([() => rule.value.transOnly, () => rule.value.transOpen], () => {
	if (
		rule.value.transOnly === 'true' &&
		rule.value.transOpen === 'true' &&
		props.element.querySelector(APP_LCNAME)
	) {
		Array.from(props.element.childNodes).forEach((el) => {
			if (el.localName !== APP_LCNAME) {
				el.remove();
			}
		});
	}
});

function getStyleByType(textStyle) {
	switch (textStyle) {
		case OPT_STYLE_LINE:
		case OPT_STYLE_DOTLINE:
		case OPT_STYLE_DASHLINE:
		case OPT_STYLE_WAVYLINE:
			return {
				opacity: '0.6',
				'-webkit-opacity': '0.6',
				'text-decoration-line': 'underline',
				'text-decoration-style': LINE_STYLES[textStyle],
				'text-decoration-color': rule.value.bgColor,
				'text-decoration-thickness': '2px',
				'text-underline-offset': '0.3em',
				'-webkit-text-decoration-line': 'underline',
				'-webkit-text-decoration-style': LINE_STYLES[textStyle],
				'-webkit-text-decoration-color': rule.value.bgColor,
				'-webkit-text-decoration-thickness': '2px',
				'-webkit-text-underline-offset': '0.3em'
			};
		case OPT_STYLE_FUZZY:
			return {
				filter: 'blur(0.2em)',
				'-webkit-filter': 'blur(0.2em)'
			};
		case OPT_STYLE_HIGHLIGHT:
			return {
				color: '#fff',
				'background-color': rule.value.bgColor || DEFAULT_COLOR
			};
		case OPT_STYLE_BLOCKQUOTE:
			return {
				opacity: '0.6',
				'-webkit-opacity': '0.6',
				display: 'block',
				padding: '0 0.75em',
				'border-left': `0.25em solid ${rule.value.bgColor || DEFAULT_COLOR}`
			};
		case OPT_STYLE_DIY:
			return rule.value.textDiyStyle;
		default:
			return {};
	}
}
</script>

<style scoped>
.styled-span[data-style='line'],
.styled-span[data-style='dotline'],
.styled-span[data-style='dashline'],
.styled-span[data-style='wavyline'] {
	&:hover {
		opacity: 1;
		-webkit-opacity: 1;
	}
}

.styled-span[data-style='fuzzy'] {
	&:hover {
		filter: none;
		-webkit-filter: none;
	}
}

.styled-span[data-style='blockquote'] {
	&:hover {
		opacity: 1;
		-webkit-opacity: 1;
	}
}
</style>
