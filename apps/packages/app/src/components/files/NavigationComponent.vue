<template>
	<div class="q-ml-sm">
		<q-btn
			icon="sym_r_arrow_back"
			flat
			dense
			:disabled="filesStore.hasBackPath(origin_id) ? false : true"
			@click="goBack"
			class="btn-no-text btn-no-border"
			:class="[
				filesStore.hasBackPath(origin_id) ? 'items-no-drag' : '',
				origin_id === FilesIdType.PAGEID ? 'btn-size-sm' : 'btn-size-xs'
			]"
			color="ink-2"
			:style="{
				pointerEvents: `${filesStore.hasBackPath(origin_id) ? 'auto' : 'none'}`
			}"
		/>
		<q-btn
			icon="sym_r_arrow_forward"
			flat
			dense
			:disabled="filesStore.hasPrevPath(origin_id) ? false : true"
			@click="goForward"
			class="btn-no-text btn-no-border"
			:class="[
				filesStore.hasBackPath(origin_id) ? 'items-no-drag' : '',
				origin_id === FilesIdType.PAGEID ? 'btn-size-sm' : 'btn-size-xs'
			]"
			color="ink-2"
			:style="{
				pointerEvents: `${filesStore.hasPrevPath(origin_id) ? 'auto' : 'none'}`
			}"
		/>
	</div>
</template>

<script setup lang="ts">
import { useFilesStore, FilesIdType } from '../../stores/files';

const filesStore = useFilesStore();

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const goBack = () => {
	filesStore.back(props.origin_id);
};

const goForward = () => {
	filesStore.previous(props.origin_id);
};
</script>

<style lang="scss" scoped></style>
