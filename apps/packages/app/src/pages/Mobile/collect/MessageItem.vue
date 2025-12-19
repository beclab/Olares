<template>
	<div
		class="message-container q-gutter-y-sm q-pa-md"
		:class="[
			data.type === 'info' ? 'bg-blue-soft' : 'bg-red-soft',
			`text-${data.type}`
		]"
	>
		<div class="row items-center q-gutter-x-sm">
			<q-icon name="sym_r_error" size="20px" />
			<span class="text-body3">{{ data.title }}</span>
		</div>
		<div class="text-body3">
			{{ data.content }}
		</div>
	</div>
</template>

<script setup lang="ts">
interface Props {
	data: {
		title: string;
		footer: string;
		content: string;
		type: string;
	};
}

defineProps<Props>();
</script>

<style></style>
