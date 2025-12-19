<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:skip="false"
		:ok="confirmText"
		:platform="deviceStore.platform"
		:cancel="useCancel ? cancelText : false"
		@onSubmit="sureAction"
	>
		<div class="row items-center justify-between" v-if="message.length > 0">
			<div
				class="text-ink-2"
				:class="{
					'message-content': hasBorder,
					'text-body3': !deviceStore.isMobile,
					'text-body3-m': deviceStore.isMobile
				}"
				v-html="message"
				:style="deviceStore.isMobile ? 'text-align: center' : ''"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { i18n } from 'src/boot/i18n';
import { useDeviceStore } from 'src/stores/settings/device';

defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	message: {
		type: String,
		required: false,
		default: ''
	},
	useCancel: {
		type: Boolean,
		required: false,
		default: true
	},
	confirmText: {
		type: String,
		required: false,
		default: i18n.global.t('ok')
	},
	cancelText: {
		type: String,
		required: false,
		default: i18n.global.t('cancel')
	},
	hasBorder: {
		type: Boolean,
		required: false,
		default: false
	},
	isReminder: {
		type: Boolean,
		required: false,
		default: false
	}
});

const CustomRef = ref();

const sureAction = () => {
	CustomRef.value.onDialogOK();
};

const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.content-root {
	padding: 16px 16px 0;
}

.userName {
	line-height: 32px;
	height: 32px;
}

.message-content {
	width: 100%;
	border: 1px solid $separator;
	padding: 8px 12px;
	border-radius: 8px;
}
</style>
