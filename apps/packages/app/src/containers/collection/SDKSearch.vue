<template>
	<div
		class="bg-background-6 row items-center justify-between sdk-serach-container"
	>
		<div class="row items-center sdk-serach-img">
			<div
				class="bg-background-1 row item-center justify-center sdk-icon-container"
				style="width: 24px; height: 24px"
			>
				<img :src="sdkIcon" alt="sdk icon" style="width: 16px" />
			</div>
			<span class="text-subtitle3 text-ink-1">{{ title }}</span>
		</div>

		<q-btn padding="4px" @click="$emit('search')">
			<q-icon name="sym_r_search" color="ink-2" size="20px" />
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import sdkIcon from 'src/assets/plugin/sdk.svg';
import { COLLECT_THEME } from 'src/constant/provide';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';
import { inject } from 'vue';

interface Props {
	title: string;
}

defineProps<Props>();

const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);
</script>

<style lang="scss" scoped>
.sdk-serach-container {
	border-radius: 8px;
	gap: 8px;
	padding: 6px 8px;
	.sdk-serach-img {
		gap: 8px;
	}
	.sdk-icon-container {
		border-radius: 4px;
	}
}
</style>
