<template>
	<q-input
		type="textarea"
		borderless
		input-class="comment-input"
		v-model="text"
		@keydown.enter="handleEnter"
	/>
</template>

<script setup lang="ts">
import { ref, defineEmits } from 'vue';

const text = ref('');
const emit = defineEmits(['update:modelValue', 'submit']);

const handleEnter = (event: KeyboardEvent) => {
	if (event.ctrlKey) {
		emit('update:modelValue', text.value + '\n');
	} else {
		// 在这里调用提交函数（假设为 submit 方法）
		event.preventDefault();
		emit('update:modelValue', text.value);
		emit('submit', text.value);
	}
};
</script>
<style lang="scss">
.comment-input {
	height: 100px !important;
	padding: 8px 12px !important;
	resize: none !important;
}
</style>
