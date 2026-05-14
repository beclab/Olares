<template>
	<div
		class="detail-btn row justify-center items-center q-mr-xs"
		@click="unBindApp()"
	>
		<q-icon size="18px" name="sym_r_link_off">
			<q-tooltip>
				{{ t('Unbind from this GPU') }}
			</q-tooltip>
		</q-icon>
	</div>
</template>

<script setup lang="ts">
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useGPUStore } from 'src/stores/settings/gpu';
import { useApplicationStore } from 'src/stores/settings/application';
import { APP_STATUS } from 'src/constant/constants';
const props = defineProps({
	app: {
		type: String,
		required: false,
		default: ''
	},
	appName: {
		type: String,
		required: false,
		default: ''
	}
});

const $q = useQuasar();
const emits = defineEmits(['unBindApp']);
const { t } = useI18n();
const gpuStore = useGPUStore();
const applicationStore = useApplicationStore();

const unBindApp = () => {
	if (
		gpuStore.gpuList.filter(
			(e) =>
				e.apps &&
				e.apps.find((app) => app.appName == props.appName) != undefined
		).length == 1
	) {
		const app = applicationStore.applications.find(
			(e) => e.name == props.appName
		);
		if (
			app &&
			app.state != APP_STATUS.STOP.COMPLETED &&
			app.state != APP_STATUS.STOP.FAILED
		) {
			$q.dialog({
				component: ReminderDialogComponent,
				componentProps: {
					message: t('Stop the app before removing it from this GPU'),
					title: t('Remove app from this GPU?', {
						app: props.app
					}),
					useCancel: false,
					confirmText: t('i_got_it')
				}
			});
			return;
		}
	}

	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('Remove app from this GPU?', {
				app: props.app
			}),
			message: t(
				"This will remove the app's access to this GPU and release its allocated resources"
			),
			useCancel: true,
			confirmText: t('confirm'),
			cancelText: t('cancel')
		}
	}).onOk(() => {
		// deleteUserSureAction();
		emits('unBindApp');
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
