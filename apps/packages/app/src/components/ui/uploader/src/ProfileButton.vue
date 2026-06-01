<template>
	<div
		ref="buttonRef"
		class="row items-center q-pa-sm"
		:class="selected ? 'bio-selected-button' : 'bio-button'"
		:style="{
			'--width': width,
			'--height': height
		}"
		@click="handleClick"
	>
		<slot />
		<q-icon
			v-if="icon"
			:size="iconSize"
			color="ink-on-brand-black"
			:name="icon"
			class="q-mr-sm"
		/>
		<span
			:class="selected ? 'bio-selected-button__label' : 'bio-button__label'"
		>
			{{ label }}
		</span>
	</div>
</template>

<script lang="ts" setup>
import {
	computed,
	getCurrentInstance,
	inject,
	onMounted,
	ref,
	watch
} from 'vue';
import { useQuasar } from 'quasar';

const $q = useQuasar();
const props = defineProps({
	icon: {
		type: String
	},
	size: {
		type: String,
		required: false
	},
	label: {
		type: String,
		required: true
	},
	width: {
		type: String,
		default: 'auto'
	},
	height: {
		type: String,
		default: 'auto'
	},
	defaultSelected: {
		type: Boolean,
		default: true
	}
});

const addChildComponent = inject('addChildComponent');
const handleItemClick = inject('handleItemClick');
const selected = ref(true);
const selfRef = ref(getCurrentInstance());

const iconSize = computed(() => {
	if (props.size) {
		return props.size;
	}
	return $q.platform.is.mobile ? '24px' : '16px';
});

onMounted(() => {
	selected.value = props.defaultSelected;
});

const setSelected = (isSelected: boolean) => {
	selected.value = isSelected;
};

defineExpose({ setSelected });

const handleClick = () => {
	if (handleItemClick) {
		handleItemClick(selfRef);
	}
};

onMounted(() => {
	if (addChildComponent) {
		addChildComponent(selfRef);
	}
});
</script>

<style scoped lang="scss">
.bio-button {
	background: $background-1;
	border: 1px solid $btn-stroke;
	height: var(--height);
	width: var(--width);
	padding-left: 12px;
	padding-right: 12px;
	justify-content: center;
	align-items: center;
	border-radius: 8px;
	cursor: pointer;
	text-decoration: none;
	color: $ink-1;

	&:hover {
		background: var(
			--profile,
			linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
		);
		color: $ink-on-brand-black;
	}

	&:active {
		background: var(
			--profile,
			linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
		);
		filter: brightness(0.9);
		color: $ink-on-brand-black;
		//transform: scale(0.98);
	}

	&__label {
		text-align: center;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}

.bio-selected-button {
	background: var(
		--profile,
		linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
	);
	height: var(--height);
	width: var(--width);
	padding-left: 12px;
	padding-right: 12px;
	justify-content: center;
	align-items: center;
	border-radius: 8px;
	cursor: pointer;
	text-decoration: none;

	&:hover {
		background: var(
			--profile,
			linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
		);
	}

	&:active {
		background: var(
			--profile,
			linear-gradient(90deg, #8ce3ff -2.75%, #7fff93 102.75%)
		);
		filter: brightness(0.9);
		//transform: scale(0.98);
	}

	&__label {
		text-align: center;
		overflow: hidden;
		color: $ink-on-brand-black;
		text-overflow: ellipsis;
	}
}
</style>
