<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="rightText"
		:cancel="leftText"
		size="small"
		platform="mobile"
		@onSubmit="onDialogOK"
	>
		<div class="content text-body2 text-ink-2">{{ content }}</div>
	</bt-custom-dialog>
</template>
<script setup lang="ts">
import { ref } from 'vue';

defineProps({
	title: String,
	content: String,
	leftText: String,
	rightText: String
});

const CustomRef = ref();

const onDialogOK = () => {
	CustomRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.content {
	text-align: center;
	margin-top: 20px;
}
</style>
