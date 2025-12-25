<template>
	<div
		:class="
			modelValue === model ? 'theme-selected' : 'theme-unselect cursor-pointer'
		"
		class="column"
		@click="emit('update:modelValue', props.model)"
	>
		<q-img class="theme-img" :src="getRequireImage(image)">
			<template v-slot:loading>
				<q-skeleton class="theme-img" />
			</template>
		</q-img>
		<div class="theme-select column justify-center">
			<bt-check-box
				:label="label"
				:circle="true"
				:model-value="modelValue === model"
				@update:model-value="updateModelValue"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from '../../utils/rss-utils';
import BtCheckBox from './BtCheckBox.vue';

const props = defineProps({
	modelValue: {
		type: String,
		required: true
	},
	model: {
		type: String,
		required: true
	},
	image: {
		type: String,
		required: true
	},
	label: {
		type: String,
		required: false
	}
});

const emit = defineEmits(['update:modelValue']);

const updateModelValue = (status: boolean) => {
	if (status) {
		emit('update:modelValue', props.model);
	}
};
</script>

<style scoped lang="scss">
.base-theme {
	overflow: hidden;
	display: flex;
	width: 210px;
	flex-direction: column;
	align-items: flex-start;
	border-radius: 12px;
	height: 148px;

	.theme-img {
		width: 100%;
		height: 100px;
	}
	.theme-select {
		width: 100%;
		height: 44px;
	}
}

.theme-unselect {
	@extend .base-theme;
	border: 2px solid $separator;
}
.theme-selected {
	@extend .base-theme;
	border: 2px solid $orange-disabled;
}
</style>
