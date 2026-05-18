<template>
	<div
		class="olares-md-preview info-content"
		:class="{
			'q-pa-md': $q.platform.is.mobile,
			'q-pa-lg': !$q.platform.is.mobile,
			'olares-dark': $q.dark.isActive,
			'olares-light': !$q.dark.isActive
		}"
		v-html="renderMarkdownToHtml(content)"
	/>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import { renderMarkdownToHtml } from 'src/utils/md';

defineProps({
	content: {
		type: String,
		default: '',
		required: false
	}
});

const $q = useQuasar();
</script>

<style scoped lang="scss">
@import '../../../../css/common/files/config/olares-markdown-preview.css';
.info-content {
	overflow-y: scroll;
	height: 100%;
}
</style>
