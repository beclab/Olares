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
const props = defineProps({
	app: {
		type: String,
		required: false,
		default: ''
	}
});

const $q = useQuasar();
const emits = defineEmits(['unBindApp']);
const { t } = useI18n();

const unBindApp = () => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			message: t('Are you sure to unbind “{app}” from this GPU?', {
				app: props.app
			}),
			title: t('Unbind App'),
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
