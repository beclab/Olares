<template>
	<div class="text-body2 text-ink-1 search-default-container">
		<MarkDownPreview
			:text="text"
			loading-hidden
			:loading="loading"
			:ended="ended"
			class="search-default-preview"
		></MarkDownPreview>
	</div>
	<OthreOperation class="gutter-top"></OthreOperation>
</template>

<script setup lang="ts">
import OthreOperation from './OthreOperation.vue';
import MarkDownPreview from './MarkDownPreview.vue';

interface Props {
	text: string;
	loading?: boolean;
	ended?: boolean;
}

withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.gutter-top {
	margin-top: 20px;
}
.search-default-container {
	.search-default-preview {
		::v-deep(pre) {
			border: 1px solid $separator;
			border-radius: 12px;
		}
		::v-deep(code:only-child) {
			width: 333px;
		}
	}
}
</style>
