<template>
	<div class="avatar-img-container" :class="[outlined ? 'avatar-outline' : '']">
		<div class="avatar-img-wrapper">
			<q-img
				class="title-img my-avatar-img"
				fit="contain"
				:width="'100%'"
				:height="'100%'"
				spinner-size="sm"
				spinner-color="grey-1"
				loading="lazy"
				:src="src"
				v-bind="$attrs"
			>
				<template #loading>
					<q-skeleton type="rect" class="full-width full-height" />
				</template>
			</q-img>
		</div>
	</div>
</template>

<script setup lang="ts">
interface Props {
	src?: string;
	outlined?: boolean;
}
const props = withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.avatar-img-container {
	position: relative;
	border-radius: 8px;
	overflow: hidden;
	width: 32px;
	height: 32px;
	box-shadow: 0px 2px 4px 0px rgba(0, 0, 0, 0.1);
	display: flex;
	align-items: stretch;
	&.avatar-outline {
		border: 2px solid $background-1;
	}

	.avatar-img-wrapper {
		flex: 1;
	}
}
</style>
