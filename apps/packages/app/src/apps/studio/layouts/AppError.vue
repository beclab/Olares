<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="$t('deploy_error')"
		:cancel="false"
		:ok="$t('btn_confirm')"
		size="medium"
		@onSubmit="submit"
	>
		<div class="status-info-container">
			<pre class="status-info-content">{{ formattedInfo }}</pre>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';

interface Props {
	errorMessage: any;
}
const CustomRef = ref();
const props = defineProps<Props>();

const formattedInfo = computed(() => {
	try {
		if (typeof props.errorMessage === 'string') {
			const parsed = JSON.parse(props.errorMessage);
			return JSON.stringify(parsed, null, 2);
		}
		return JSON.stringify(props.errorMessage, null, 2);
	} catch (error) {
		return props.errorMessage;
	}
});
const submit = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style lang="scss" scoped>
.status-info-container {
	max-height: 60vh;
	overflow-y: auto;
	background: $background-2;
	border-radius: 8px;
	border: 1px solid $separator;
}

.status-info-content {
	padding: 16px;
	margin: 0;
	font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Courier New', monospace;
	font-size: 13px;
	line-height: 1.6;
	color: $ink-2;
	white-space: pre-wrap;
	word-break: break-word;
	overflow-x: auto;
}
</style>
