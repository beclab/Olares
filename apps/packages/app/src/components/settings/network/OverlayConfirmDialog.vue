<template>
	<bt-custom-dialog
		ref="customRef"
		:title="dialogTitle"
		:skip="false"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="small"
		:platform="deviceStore.platform"
		@onSubmit="onConfirm"
	>
		<div class="overlay-confirm-dialog__body">
			<i18n-t
				:keypath="messageKey"
				tag="div"
				scope="global"
				class="text-ink-2 overlay-confirm-dialog__message"
				:class="{
					'text-body3': !deviceStore.isMobile,
					'text-body3-m': deviceStore.isMobile
				}"
			>
				<template #appTitle>
					<strong>{{ props.appTitle }}</strong>
				</template>
			</i18n-t>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';

const props = defineProps<{
	action: 'enable' | 'disable';
	appTitle: string;
	isStopped?: boolean;
}>();

const { t } = useI18n();
const deviceStore = useDeviceStore();
const customRef = ref();

const MSG_ENABLE_OVERLAY_CONFIRM =
	'Are you sure you want to enable overlay for {appTitle}?\n\nThe application will restart immediately and become accessible via IP once running.';
const MSG_DISABLE_OVERLAY_CONFIRM =
	'Are you sure you want to disable overlay for {appTitle}?\n\nThe application will restart immediately and become inaccessible via IP.';
// Stopped apps are not restarted, so the copy drops the "restart immediately" wording.
const MSG_ENABLE_OVERLAY_STOPPED_CONFIRM =
	'Are you sure you want to enable overlay for {appTitle}?\n\nThe application will become accessible via IP once running.';

const dialogTitle = computed(() =>
	props.action === 'enable'
		? t('Enable overlay', { appTitle: props.appTitle })
		: t('Disable overlay', { appTitle: props.appTitle })
);

const messageKey = computed(() =>
	props.action === 'enable'
		? props.isStopped
			? MSG_ENABLE_OVERLAY_STOPPED_CONFIRM
			: MSG_ENABLE_OVERLAY_CONFIRM
		: MSG_DISABLE_OVERLAY_CONFIRM
);

const onConfirm = () => {
	customRef.value?.onDialogOK();
};
</script>

<style scoped lang="scss">
.overlay-confirm-dialog__body {
	padding: 0 4px;
	word-wrap: break-word;
}

.overlay-confirm-dialog__message {
	white-space: pre-wrap;
}
</style>
