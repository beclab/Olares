<template>
	<div class="text-center" @click="clickHandler">
		<q-img :src="emptyIcon" :ratio="1" :width="imgSize" spinner-size="0px" />
		<div class="column no-wrap flex-gap-y-xs">
			<div class="text-subtitle2 text-ink-1">
				{{ title }}
			</div>
			<div class="text-body3 text-ink-3">
				<span v-if="subtitle">
					{{ subtitle }}
				</span>
				<slot v-else-if="$slots.subtitle" name="subtitle"></slot>
			</div>
		</div>
		<div class="q-mt-lg" v-if="!btnHidden">
			<CustomButton
				:label="$t('bex.try_again')"
				color="yellow-default"
				text-color="ink-on-brand-black"
				class="q-px-xxl"
			></CustomButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import emptyIcon from 'src/assets/plugin/empty.svg';
import CustomButton from './CustomButton.vue';
import { computed } from 'vue';

enum Size {
	sm = '120px',
	md = '140px'
}

interface Props {
	title: string;
	subtitle?: string;
	btnHidden?: boolean;
	size?: keyof typeof Size;
}

const props = withDefaults(defineProps<Props>(), {
	btnHidden: false,
	size: 'md'
});

const imgSize = computed(() => Size[props.size]);

const emit = defineEmits(['click']);

const clickHandler = () => {
	emit('click');
};
</script>

<style></style>
