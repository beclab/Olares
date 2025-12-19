<template>
	<div
		v-if="skeleton"
		:style="{
			'--size': `${size}px`,
			'--appSize': `${size - 8}px`,
			'--csSize': `${csSize}px`,
			'--iconRadius': `${iconRadius}px`
		}"
		class="application_shadow row justify-center items-center"
	>
		<q-skeleton class="application_icon" />
	</div>

	<div
		v-else
		:style="{
			'--size': `${size}px`,
			'--appSize': `${size - 8}px`,
			'--csSize': `${csSize}px`,
			'--iconRadius': `${iconRadius}px`
		}"
		class="application_shadow row justify-center items-center"
	>
		<q-img :src="src" class="application_icon">
			<template v-slot:loading>
				<q-skeleton class="application_icon" />
			</template>
		</q-img>
		<q-img
			v-if="csApp"
			class="application_cs_icon"
			:src="getRequireImage('share_app.svg')"
		/>
	</div>
</template>

<script lang="ts" setup>
import { getRequireImage } from '@apps/market/src/utils/imageUtils';

defineProps({
	src: {
		type: String,
		require: true
	},
	csApp: {
		type: Boolean,
		default: false
	},
	skeleton: {
		type: Boolean,
		default: false
	},
	size: {
		type: Number,
		default: 64
	},
	csSize: {
		type: Number,
		default: 20
	},
	iconRadius: {
		type: Number,
		default: 12
	}
});
</script>

<style scoped lang="scss">
.application_shadow {
	width: var(--size);
	height: var(--size);
	justify-content: center;
	align-items: center;
	position: relative;

	.application_icon {
		width: var(--appSize);
		height: var(--appSize);
		box-shadow: 0 2px 4px 0 #0000001a;
		border-radius: var(--iconRadius);
		overflow: hidden;
	}

	.application_cs_icon {
		background: transparent;
		width: var(--csSize);
		height: var(--csSize);
		position: absolute;
		filter: drop-shadow(0px 2px 4px rgba(0, 0, 0, 0.2));
		right: 0;
		bottom: 0;
		z-index: 3;
	}
}
</style>
