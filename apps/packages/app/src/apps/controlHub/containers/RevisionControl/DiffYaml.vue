<template>
	<div class="q-my-md row items-center">
		<q-icon size="xs" name="terminal" />
		<span class="q-ml-xs">{{ t('CONFIG_FILE') }}</span>

		<q-icon
			size="sm"
			name="expand_more"
			:color="mode === 'line-by-line' ? 'primary' : ''"
			@click="() => modeChange('line-by-line')"
		>
			<q-tooltip anchor="top middle" self="bottom middle">
				line-by-line
			</q-tooltip>
		</q-icon>
		<q-icon
			size="sm"
			name="navigate_next"
			:color="mode === 'side-by-side' ? 'primary' : ''"
			@click="() => modeChange('side-by-side')"
		>
			<q-tooltip anchor="top middle" self="bottom middle">
				side-by-side
			</q-tooltip>
		</q-icon>
	</div>
	<div class="diff-yaml-wrapper" v-html="diffHtml"></div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

import { createPatch } from 'diff';
import { parse, html, Diff2HtmlConfig } from 'diff2html';
import { t } from '@apps/control-hub/src/boot/i18n';
import 'diff2html/bundles/css/diff2html.min.css';
import { useQuasar } from 'quasar';

const $q = useQuasar();

interface Props {
	datas: any;
}

const props = withDefaults(defineProps<Props>(), {});

const mode = ref('side-by-side');
const getDiffHtml = ([oldData, newData]: any, options = {}) => {
	const diffStr = createPatch('yaml', oldData, newData, '', '', {
		context: -1
	});
	const diffJson = parse(diffStr);
	return html(diffJson, {
		drawFileList: false,
		colorScheme: $q.dark.isActive ? 'dark' : 'light',
		...options
	} as Diff2HtmlConfig & { colorScheme: 'dark' | 'light'; auto });
};

const diffHtml = computed(() =>
	props.datas ? getDiffHtml(props.datas, { outputFormat: mode.value }) : ''
);

const modeChange = (type: string) => {
	mode.value = type;
};
</script>
