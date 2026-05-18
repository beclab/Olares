<template>
	<div class="text-center" @click="clickHandler">
		<q-img
			:src="emptyIcon || emptyIconDefault"
			:ratio="1"
			:width="imgSize"
			spinner-size="0px"
		/>
		<div class="column items-center no-wrap flex-gap-y-xs">
			<div class="text-subtitle2 text-ink-1">
				{{ title }}
			</div>
			<div class="text-body3 text-ink-3">
				<span v-if="subtitle">
					{{ subtitle }}
				</span>
				<slot v-else-if="$slots.subtitle" name="subtitle"></slot>
			</div>
			<ul v-if="listItems && listItems.length > 0" class="empty-list">
				<li
					v-for="(item, index) in listItems"
					:key="index"
					class="text-body3 text-ink-3 text-left"
				>
					{{ item }}
				</li>
			</ul>
		</div>
		<slot name="action" v-if="$slots.action"></slot>
		<div class="q-mt-lg" v-else-if="!btnHidden">
			<CustomButton
				:label="btnLabel || $t('bex.try_again')"
				color="yellow-default"
				icon="sym_r_autorenew"
				text-color="ink-on-brand-black"
				class="q-px-xxl"
			></CustomButton>
		</div>
	</div>
</template>

<script setup lang="ts">
import emptyIconDefault from 'src/assets/plugin/empty.svg';
import CustomButton from './CustomButton.vue';
import { computed } from 'vue';

enum Size {
	sm = '120px',
	md = '140px'
}

interface Props {
	title: string;
	subtitle?: string;
	listItems?: string[];
	btnHidden?: boolean;
	btnLabel?: string;
	size?: keyof typeof Size;
	emptyIcon?: string;
}

const props = withDefaults(defineProps<Props>(), {
	btnHidden: false,
	btnLabel: '',
	size: 'md'
});

const imgSize = computed(() => Size[props.size]);

const emit = defineEmits(['click']);

const clickHandler = () => {
	emit('click');
};
</script>

<style scoped>
.empty-list {
	list-style: none;
	padding: 0;
	margin: 4px 0 0 0;
	display: inline-block;
}

.empty-list li {
	position: relative;
	padding-left: 10px;
	margin-bottom: 4px;
}

.empty-list li::before {
	content: '•';
	position: absolute;
	left: 0;
	color: currentColor;
}
</style>
