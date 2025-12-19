<template>
	<div
		:style="{ width: size, height: size }"
		style="position: relative; background: transparent"
		class="row justify-center items-center"
	>
		<div style="width: 90%; height: 90%; border-radius: 50%">
			<slot />
		</div>
		<q-img
			v-if="showBorder"
			:style="{ width: size, height: size }"
			style="position: absolute; top: 0"
			src="https://app.cdn.olares.com/avatar3/nft.png"
		/>
	</div>
</template>
<script setup lang="ts">
defineProps({
	size: {
		type: String,
		require: true
	},
	showBorder: {
		type: Boolean,
		default: true
	}
});
</script>
