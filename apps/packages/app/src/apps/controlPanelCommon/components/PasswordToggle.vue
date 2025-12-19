<template>
	<div class="row inline items-center">
		<span v-if="visible">{{ value }}</span>
		<div class="dot-container row items-center text-ink-2" v-else>
			<div v-for="item in 8" :key="item" class="dot row items-center"></div>
		</div>
		<q-icon
			:name="visible ? 'sym_r_visibility_off' : 'sym_r_visibility'"
			class="q-ml-sm cursor-pointer"
			size="20px"
			color="ink-3"
			@click="() => (visible = !visible)"
		/>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

interface Props {
	value: string;
}

const visible = ref(false);

withDefaults(defineProps<Props>(), {});
</script>

<style lang="scss" scoped>
.dot-container {
	display: flex;
	gap: 4px;
	flex-wrap: nowrap;
	.dot {
		height: 8px;
		font-size: 16px;
		width: 8px;
		border-radius: 50%;
		background: $ink-2;
		align-self: baseline;
	}
}
</style>
