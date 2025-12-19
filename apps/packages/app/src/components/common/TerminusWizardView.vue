<template>
	<div class="wizard-view-root">
		<q-scroll-area
			style="width: 100%"
			:style="`height: calc(100% - 96px - ${bottomMoreHeight}px)`"
			content-style="width:100%;height:100%"
		>
			<div class="wizard-view-root-content">
				<slot name="content"></slot>
			</div>
		</q-scroll-area>
		<div class="bottom-view column justify-end">
			<slot name="bottom-top"></slot>
			<confirm-button
				v-if="enableBtn"
				:btn-title="btnTitle"
				:btn-icon="btnIcon"
				@onConfirm="onConfirm"
				@onError="onError"
				:btn-status="btnStatus"
			/>
			<slot name="bottom-bottom"></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import ConfirmButton from './ConfirmButton.vue';
import { onMounted, onBeforeUnmount } from 'vue';
import { useQuasar } from 'quasar';
import { i18n } from '../../boot/i18n';
import { ConfirmButtonStatus } from '../../utils/constants';
import { useMenuStore } from '../../stores/menu';
import AndroidPlugins from 'src/platform/interface/capacitor/android/androidPlugins';

const props = defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	marginLeft: {
		type: Number,
		required: false,
		default: 12
	},
	btnTitle: {
		type: String,
		required: false,
		default: i18n.global.t('confirm')
	},
	btnStatus: {
		type: Number,
		require: false,
		default: ConfirmButtonStatus.normal
	},
	btnIcon: {
		type: String,
		required: false,
		default: ''
	},
	enableOverlay: {
		type: Boolean,
		required: false,
		default: true
	},
	bottomMoreHeight: {
		type: Number,
		default: 0,
		required: false
	},
	enableBtn: {
		type: Boolean,
		required: false,
		default: true
	}
});

const $q = useQuasar();

const menuStore = useMenuStore();

onMounted(async () => {
	if (props.enableOverlay) {
		if ($q.platform.is.android) {
			// StatusBar.setOverlaysWebView({ overlay: true });
			await AndroidPlugins.EdgeToEdge.disable();
		}
		menuStore.changeSafeArea(false);
	}
});

onBeforeUnmount(async () => {
	if (props.enableOverlay) {
		if ($q.platform.is.android) {
			// StatusBar.setOverlaysWebView({ overlay: false });
			await AndroidPlugins.EdgeToEdge.enable();
		}
		menuStore.changeSafeArea(true);
	}
});

const emit = defineEmits(['onConfirm', 'onError']);

const onConfirm = () => {
	emit('onConfirm');
};

const onError = () => {
	emit('onError');
};
</script>

<style scoped lang="scss">
.wizard-view-root {
	width: 100%;
	height: 100%;
	padding-left: 20px;
	padding-right: 20px;

	.wizard-view-root-content {
		width: 100%;
		height: 100%;
	}

	.bottom-view {
		width: 100%;
		padding-bottom: 48px;
	}
}
</style>
