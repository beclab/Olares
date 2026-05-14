<template>
	<div class="q-ml-sm">
		<q-btn
			icon="sym_r_arrow_back"
			flat
			dense
			:disabled="!canGoBack"
			@click="goBack"
			class="btn-no-text btn-no-border"
			:class="[
				canGoBack ? 'items-no-drag' : '',
				origin_id === FilesIdType.PAGEID ? 'btn-size-sm' : 'btn-size-xs'
			]"
			color="ink-2"
			:style="{ pointerEvents: canGoBack ? 'auto' : 'none' }"
		/>
		<q-btn
			icon="sym_r_arrow_forward"
			flat
			dense
			:disabled="!canGoForward"
			@click="goForward"
			class="btn-no-text btn-no-border"
			:class="[
				canGoForward ? 'items-no-drag' : '',
				origin_id === FilesIdType.PAGEID ? 'btn-size-sm' : 'btn-size-xs'
			]"
			color="ink-2"
			:style="{ pointerEvents: canGoForward ? 'auto' : 'none' }"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useFilesStore, FilesIdType } from '../../stores/files';

const filesStore = useFilesStore();

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const canGoBack = computed(
	() =>
		!filesStore.backPathIsHome(props.origin_id) &&
		filesStore.hasBackPath(props.origin_id)
);

const canGoForward = computed(() => filesStore.hasPrevPath(props.origin_id));

const goBack = () => {
	filesStore.back(props.origin_id);
};

const goForward = () => {
	filesStore.previous(props.origin_id);
};
</script>

<style lang="scss" scoped></style>
