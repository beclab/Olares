<template>
	<div class="row justify-end">
		<div
			class="my-query-item q-py-md q-px-lg text-left text-body2 text-grey-10 bg-yellow-6"
		>
			<span v-html="format(text)"></span>
		</div>
	</div>
</template>

<script setup lang="ts">
import { format } from '../utils';
interface Props {
	text: string;
}
withDefaults(defineProps<Props>(), {});
</script>
<style lang="scss" scoped>
.my-query-item {
	display: inline-block;
	border-top-left-radius: 20px;
	border-bottom-left-radius: 20px;
	border-top-right-radius: 20px;
	border-bottom-right-radius: 4px;
	word-break: break-all;
}
</style>
