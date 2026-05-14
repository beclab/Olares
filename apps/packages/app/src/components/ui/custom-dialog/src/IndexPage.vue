<template>
	<q-dialog
		class="card-dialog"
		ref="dialogRef"
		v-model="show"
		:position="position"
		:persistent="persistent"
		@hide="hiddenDialog"
		:noRouteDismiss="noRouteDismiss"
	>
		<q-card
			ref="cardRef"
			class="card-container no-shadow column no-wrap"
			:class="[paddingC, { resizable: resizable }]"
			:style="{
				width: resolvedWidth,
				maxWidth: resizable ? 'none' : resolvedWidth,
				height: resolvedHeight,
				...(resizable ? { resize: 'both', overflow: 'hidden' } : {})
			}"
		>
			<template v-if="$slots.header">
				<slot name="header"></slot>
			</template>
			<template v-else>
				<dialog-bar
					:title="title"
					:icon="icon"
					:platform="platform"
					@close="onCancel"
				/>
			</template>

			<div class="dialog-content">
				<slot />
			</div>

			<dialog-footer
				:ok="ok"
				:cancel="cancel"
				:okStyle="okStyle"
				:loading="okLoading"
				:platform="platform"
				:skip="skip"
				:okDisabled="okDisabled"
				:okClass="okClass"
				:disableCancelFucus="disableCancelFucus"
				@onCancel="onCancel"
				@onSubmit="onSubmit"
				@onSkip="onSkip"
			>
				<template #footerMore>
					<slot name="footerMore" />
				</template>
			</dialog-footer>
		</q-card>
	</q-dialog>
</template>

<script lang="ts" setup>
import { useDialogPluginComponent } from 'quasar';
import {
	ref,
	defineProps,
	computed,
	watch,
	onMounted,
	onUnmounted,
	nextTick
} from 'vue';

import DialogBar from './DialogBar.vue';
import DialogFooter from './DialogFooter.vue';

import { Platform, Size } from './type';

interface Props {
	platform: Platform;
	size: Size;
	title: string;
	icon: string;
	persistent: boolean;
	ok: string | boolean;
	okStyle: object;
	okClass: string;
	cancel: string | boolean;
	okLoading: string | boolean;
	skip: string | boolean;
	fullWidth: boolean;
	fullHeight: boolean;
	okDisabled: boolean;
	noRouteDismiss: boolean;
	modelValue?: boolean;
	barRedefined?: boolean;
	cancelDismiss: boolean;
	position: 'standard' | 'top' | 'right' | 'bottom' | 'left';
	contentPending: boolean;
	disableCancelFucus: boolean;
	resizable: boolean;
	resizableHeight: string;
	resizableWidth: string;
	storageKey: string;
}

const props = withDefaults(defineProps<Props>(), {
	platform: Platform.WEB,
	size: Size.SMALL,
	persistent: false,
	ok: true,
	okStyle: () => ({}),
	okClass: '',
	okLoading: false,
	fullWidth: false,
	fullHeight: false,
	okDisabled: false,
	noRouteDismiss: false,
	modelValue: true,
	barRedefined: false,
	cancelDismiss: true,
	position: 'standard',
	contentPending: true,
	disableCancelFucus: false,
	resizable: false,
	resizableHeight: '600px',
	resizableWidth: '',
	storageKey: ''
});

const emits = defineEmits([
	'onSubmit',
	'onCancel',
	'onSkip',
	'onHide',
	'update:modelValue'
]);

const { dialogRef, onDialogCancel, onDialogOK, onDialogHide } =
	useDialogPluginComponent();

console.log(props.modelValue);

const show = ref(props.modelValue);

watch(
	() => props.modelValue,
	(newValue) => {
		show.value = newValue;
	}
);

const widthRatio = ref(0.86);
const heightRatio = ref(0.75);

const width = computed(() => {
	if (props.fullWidth) {
		const innerWidth = window.innerWidth;
		return innerWidth * widthRatio.value + 'px';
	}

	switch (props.size) {
		case Size.SMALL:
			return '400px';

		case Size.MEDIUM:
			return '560px';

		case Size.LARGE:
			return '800px';

		default:
			return '400px';
	}
});

const height = computed(() => {
	if (props.fullHeight) {
		const innerHeight = window.innerHeight;
		return innerHeight * heightRatio.value + 'px';
	}

	if (props.resizable) {
		return props.resizableHeight;
	}

	return 'auto';
});

const cardRef = ref<any>(null);
const savedWidth = ref('');
const savedHeight = ref('');

const resolvedWidth = computed(() => {
	if (props.resizable && savedWidth.value) return savedWidth.value;
	if (props.resizable && props.resizableWidth) return props.resizableWidth;
	return width.value;
});

const resolvedHeight = computed(() =>
	props.resizable && savedHeight.value ? savedHeight.value : height.value
);

let _roDisconnect: (() => void) | null = null;
let _saveTimer: ReturnType<typeof setTimeout> | null = null;

onMounted(() => {
	if (!props.resizable || !props.storageKey) return;

	try {
		const raw = localStorage.getItem(props.storageKey);
		if (raw) {
			const { w, h } = JSON.parse(raw) as { w: string; h: string };
			if (w) savedWidth.value = w;
			if (h) savedHeight.value = h;
		}
	} catch (e) {
		console.error(e);
	}

	nextTick(() => {
		const el: HTMLElement | undefined = cardRef.value?.$el;
		if (!el) return;
		const ro = new ResizeObserver(() => {
			const newW = `${el.offsetWidth}px`;
			const newH = `${el.offsetHeight}px`;

			if (savedWidth.value !== newW) savedWidth.value = newW;
			if (savedHeight.value !== newH) savedHeight.value = newH;

			if (_saveTimer) clearTimeout(_saveTimer);
			_saveTimer = setTimeout(() => {
				try {
					localStorage.setItem(
						props.storageKey,
						JSON.stringify({ w: newW, h: newH })
					);
				} catch (e) {
					console.error(e);
				}
			}, 300);
		});
		ro.observe(el);
		_roDisconnect = () => ro.disconnect();
	});
});

onUnmounted(() => {
	_roDisconnect?.();
	if (_saveTimer) clearTimeout(_saveTimer);
});

const onSubmit = async () => {
	emits('onSubmit');
};

let hidden = true;

const onCancel = () => {
	hidden = false;
	emits('onCancel');
	if (props.cancelDismiss) {
		onDialogCancel();
	}
};

const hiddenDialog = () => {
	if (hidden) {
		emits('onHide');
	}
	emits('update:modelValue', false);
};

const onSkip = () => {
	emits('onSkip');
};

const paddingC = computed(() => {
	if (props.contentPending) {
		if (props.position == 'bottom') {
			return 'position-bottom normal-bottom-pending';
		}
		return 'normal-pending';
	}

	if (props.position == 'bottom') {
		return 'position-bottom normal-disable-bottom-pending';
	}
	return 'normal-disable-pending';
});

defineExpose({
	onDialogOK,
	onDialogCancel,
	onDialogHide
});
</script>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
	name: 'BtCustomDialog'
});
</script>

<style lang="scss" scoped>
.card-dialog {
	.card-container {
		border-radius: 12px;

		.dialog-content {
			flex: 1;
			margin: 20px 0 32px;
			width: 100%;
			max-height: calc(100vh * 0.75);
			overflow: scroll;

			&::-webkit-scrollbar {
				display: none;
			}
		}
	}

	.position-bottom {
		position: fixed;
		bottom: 0;
		padding-bottom: env(safe-area-inset-bottom);
	}

	.normal-pending {
		padding: 20px;
	}

	.normal-bottom-pending {
		padding: 20px;
		padding-bottom: calc(env(safe-area-inset-bottom) + 20px);
	}

	.normal-disable-pending {
		padding: 0;
	}

	.normal-disable-bottom-pending {
		padding: 0;
		padding-bottom: env(safe-area-inset-bottom);
	}

	.card-container.resizable {
		.dialog-content {
			flex: 1;
			max-height: none;
			overflow: hidden;
		}
	}
}
</style>
