<template>
	<div
		class="detail-btn row justify-center items-center q-mr-xs"
		@click="unBindApp()"
	>
		<q-icon size="18px" name="sym_r_repeat">
			<q-tooltip>
				{{ t('Switch GPU') }}
			</q-tooltip>
		</q-icon>
	</div>
</template>

<script setup lang="ts">
import SwitchGPUDialog from './SwitchGPUDialog.vue';
import { useQuasar } from 'quasar';
import { GPUInfo } from 'src/stores/settings/gpu';
import { PropType } from 'vue';
import { useI18n } from 'vue-i18n';
const props = defineProps({
	app: {
		type: String,
		required: false,
		default: ''
	},
	currentGPU: {
		type: Object as PropType<GPUInfo>,
		required: true
	},
	appName: {
		type: String,
		required: false,
		default: ''
	}
});

const $q = useQuasar();
const { t } = useI18n();

const unBindApp = () => {
	$q.dialog({
		component: SwitchGPUDialog,
		componentProps: {
			// message: t('Are you sure to unbind “{app}” from this GPU?', {
			// 	app: props.app
			// }),
			// title: t('Unbind App'),
			// useCancel: true,
			// confirmText: t('confirm'),
			// cancelText: t('cancel')
			currentGPU: props.currentGPU,
			appName: props.appName,
			appTitle: props.app
		}
	}).onOk(() => {
		// deleteUserSureAction();
		// emits('unBindApp');
	});
};
</script>

<style scoped lang="scss">
.detail-btn {
	cursor: pointer;
	height: 24px;
	width: 24px;
	color: $ink-2;
}
</style>
