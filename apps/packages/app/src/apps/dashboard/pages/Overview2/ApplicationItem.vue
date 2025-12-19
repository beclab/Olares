<template>
	<div class="row items-center justify-between q-py-sm">
		<div class="row items-center">
			<div class="text-subtitle1 text-ink-1" style="min-width: 20px">
				<q-skeleton v-if="loading" type="text" />
				<span v-else>{{ index }}</span>
			</div>
			<div class="q-ml-lg q-mr-md">
				<MyImg
					:src="data.icon"
					width="36px"
					height="36px"
					:loading="loading"
				></MyImg>
			</div>
			<div class="text-subtitle1 text-ink-1">
				<q-skeleton v-if="loading" type="text" width="72px" />
				<span v-else>{{ data.title }}</span>
			</div>
		</div>
		<div class="text-subtitle1 text-ink-3">
			<q-skeleton v-if="loading" type="text" width="48px" />
			<span v-else>{{ data.used }} {{ data.unit }}</span>
		</div>
	</div>
</template>

<script setup lang="ts">
import MyImg from '@apps/control-panel-common/src/components/MyImg.vue';

interface Props {
	data: any;
	index: number;
	loading?: boolean;
}

withDefaults(defineProps<Props>(), {});
</script>

<style></style>
