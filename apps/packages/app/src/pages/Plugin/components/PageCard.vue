<template>
	<div class="bex-page-card column no-wrap full-height">
		<div class="row items-center justify-between no-wrap q-pa-lg" v-if="title">
			<span class="text-h4 text-ink-1">{{ title }}</span>
			<div v-if="$slots.extra">
				<slot name="extra"></slot>
			</div>
		</div>
		<div style="flex: 1">
			<bt-scroll-area
				style="height: 100%; width: 100%"
				class="custom-scroll-wrapper"
			>
				<div class="q-px-lg q-pb-lg">
					<slot></slot>
				</div>
			</bt-scroll-area>
			<q-inner-loading :showing="loading"> </q-inner-loading>
		</div>
	</div>
</template>

<script setup lang="ts">
interface Props {
	title?: string;
	loading?: boolean;
}

defineProps<Props>();
</script>

<style lang="scss" scoped>
.bex-page-card {
	::v-deep(.custom-scroll-wrapper .q-scrollarea__content) {
		width: 100%;
	}
}
</style>
