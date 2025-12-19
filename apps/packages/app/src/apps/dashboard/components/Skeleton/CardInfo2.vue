<template>
	<div class="my-card-container">
		<div class="row items-center">
			<div class="row items-baseline">
				<div class="text-h5">
					<q-skeleton width="32px" />
				</div>
			</div>
			<div class="q-ml-sm">
				<q-skeleton class="rounded-borders-lg" width="40px" height="20px" />
			</div>
		</div>
		<div class="text-subtitle3 text-grey-8 q-mt-sm">
			<q-skeleton width="90px" />
		</div>
	</div>
</template>
<script setup lang="ts">
interface Props {
	loading?: boolean;
}

const props = withDefaults(defineProps<Props>(), {});
</script>
<style lang="scss" scoped>
.q-badge-wrapper {
	background: $separator-color;
}
</style>
