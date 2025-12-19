<template>
	<div
		class="row justify-center items-center"
		:class="selected ? 'avatar-selector-selected' : 'avatar-selector-normal'"
	>
		<div class="selector-avatar" style="position: relative">
			<img
				v-if="src"
				:src="src"
				loading="lazy"
				style="width: 100%; height: 100%"
				@error.once="
					(e) => {
						e.target.src = 'account/nft_default_img.png';
					}
				"
			/>
			<div
				class="text-caption row items-center justify-center bound-status"
				v-if="bound"
			>
				{{ $t('bound') }}
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
defineProps({
	src: String,
	selected: {
		type: Boolean,
		default: false
	},
	bound: {
		type: Boolean,
		default: false
	}
});
</script>

<style scoped lang="scss">
.avatar-selector-selected {
	width: 104px;
	height: 104px;
	border-radius: 4px;
	border: 1px solid $blue;
}

.avatar-selector-normal {
	width: 104px;
	height: 104px;
	border-radius: 4px;
	border: none;
}

.selector-avatar {
	width: 96px;
	height: 96px;
	border-radius: 4px;
	background-color: $background-3;
	overflow: hidden;
}

.bound-status {
	width: 45px;
	height: 20px;
	bottom: 2px;
	right: 0px;
	padding: 4px, 8px, 4px, 8px;
	border-radius: 4px;
	gap: 10px;
	position: absolute;
	background-color: $grey-5;
	color: #fff;
}
</style>
