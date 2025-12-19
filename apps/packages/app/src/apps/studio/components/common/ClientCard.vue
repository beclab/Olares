<template>
	<div class="resources-item row items-center justify-between">
		<div class="resources-left row items-center justify-start">
			<img :src="dynamicImageSrc" />
			<div>
				<div class="text-ink-1 text-subtitle3">{{ data.label }}</div>
				<div class="text-ink-2 text-overline">{{ data.url }}</div>
			</div>
		</div>
		<div class="resources-right row items-center justify-end">
			<q-icon
				class="cursor-pointer"
				name="sym_r_edit_square"
				size="16px"
				color="ink-2"
				@click="editClient"
			/>
			<q-icon
				class="cursor-pointer q-ml-md"
				name="sym_r_delete"
				size="16px"
				color="ink-2"
				@click="deleteClient"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { defineProps, defineEmits, computed } from 'vue';
const props = defineProps({
	data: {
		type: Object,
		required: false,
		default: () => ({})
	}
});

const emit = defineEmits(['editClient', 'deleteClient']);

const editClient = () => {
	emit('editClient', props.data);
};

const deleteClient = () => {
	emit('deleteClient', props.data);
};

const dynamicImageSrc = computed(() => {
	return `studio/client/${props.data.name}.svg`;
});
</script>

<style lang="scss" scoped>
.resources-item {
	border: 1px solid $separator;
	border-radius: 8px;
	margin-top: 12px;
	padding: 14px 20px;
	.resources-left {
		img {
			width: 24px;
			height: 24px;
			margin-right: 12px;
		}
	}
}
</style>
