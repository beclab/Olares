<template>
	<div
		class="column inline justify-center items-center flex-gap-y-lg"
		v-if="showing"
	>
		<div
			class="spinner-loading-container"
			:style="{
				width: size,
				height: size
			}"
		>
			<img :src="icon" class="full-width" alt="loading" />
		</div>
		<div v-if="desc" class="text-center">{{ desc }}</div>
	</div>
</template>

<script setup lang="ts">
import spinnerLoaders from 'src/assets/common/spinner-loaders.svg';
import spinnerLoaders2 from 'src/assets/plugin/spinner-loading.svg';
import { computed } from 'vue';

interface Props {
	desc?: string;
	size?: string;
	showing?: boolean;
	type?: 'default' | 'overlay';
}

const props = withDefaults(defineProps<Props>(), {
	size: '20px',
	showing: true,
	type: 'default'
});

const icon = computed(() => {
	return props.type === 'overlay' ? spinnerLoaders2 : spinnerLoaders;
});
</script>

<style lang="scss" scoped>
.spinner-loading-container {
	display: inline-block;
	overflow: hidden;
	border-radius: 50%;
	animation: rotate 1.5s linear infinite;
}

@keyframes rotate {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(360deg);
	}
}
</style>
