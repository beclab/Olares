<template>
	<q-dialog
		position="bottom"
		ref="dialogRef"
		transition-show="jump-up"
		transition-hide="jump-down"
		transition-duration="300"
	>
		<terminus-dialog-display-content :dialog-ref="dialogRef">
			<template v-slot:title>
				<div class="full-height column justify-center">
					<div class="text-ink-1 text-h5">{{ appName }}</div>
				</div>
			</template>
			<template v-slot:content>
				<q-list style="margin: 8px">
					<operation-item
						v-if="showRetry"
						icon="sym_r_replay"
						:label="t('app.retry')"
						@on-item-click="
							() => {
								onRetry();
								onDialogOK();
							}
						"
					/>
					<operation-item
						v-if="showResume"
						icon="sym_r_resume"
						:label="t('app.resume')"
						@on-item-click="
							() => {
								onResume();
								onDialogOK();
							}
						"
					/>
					<operation-item
						v-if="showStop"
						icon="sym_r_stop_circle"
						:label="t('app.stop')"
						@on-item-click="
							() => {
								onStop();
								onDialogOK();
							}
						"
					/>
					<operation-item
						v-if="showOpenInUpgrade"
						icon="sym_r_open_in_browser"
						:label="t('app.open')"
						@on-item-click="
							() => {
								onUpdateOpen();
								onDialogOK();
							}
						"
					/>
					<operation-item
						v-if="showClone"
						icon="sym_r_open_in_browser"
						:label="t('app.clone')"
						@on-item-click="
							() => {
								onClone();
								onDialogOK();
							}
						"
					/>
					<operation-item
						v-if="showUninstall"
						icon="sym_r_delete_forever"
						:label="t('app.uninstall')"
						@on-item-click="
							() => {
								onUninstall();
								onDialogOK();
							}
						"
					/>
					<operation-item
						v-if="showRemoveLocal"
						icon="sym_r_do_not_disturb_on"
						:label="t('app.remove')"
						@on-item-click="
							() => {
								onRemoveLocal();
								onDialogOK();
							}
						"
					/>
				</q-list>
			</template>
		</terminus-dialog-display-content>
	</q-dialog>
</template>

<script setup lang="ts">
import TerminusDialogDisplayContent from '../common/TerminusDialogDisplayContent.vue';
import { AppStatusLatest } from 'src/constant/constants';
import { useDialogPluginComponent } from 'quasar';
import OperationItem from './OperationItem.vue';
import useAppAction from './useAppAction';
import { PropType } from 'vue';
import { useI18n } from 'vue-i18n';

const { onDialogOK, dialogRef } = useDialogPluginComponent();

const props = defineProps({
	item: {
		type: Object as PropType<AppStatusLatest>,
		required: true
	},
	appName: {
		type: String,
		required: true
	},
	version: {
		type: String,
		required: true
	},
	sourceId: {
		type: String,
		required: true
	},
	larger: {
		type: Boolean,
		required: false,
		default: false
	},
	manager: {
		type: Boolean,
		require: false,
		default: false
	}
});

const {
	showRemoveLocal,
	onRemoveLocal,
	showUninstall,
	onUninstall,
	showOpenInUpgrade,
	onUpdateOpen,
	showRetry,
	onRetry,
	showStop,
	onStop,
	showResume,
	onResume,
	showClone,
	onClone
} = useAppAction(props);

const { t } = useI18n();
</script>

<style lang="scss" scoped>
.operate-dialog-title-module {
	width: 100%;
	height: 100%;

	.title {
		text-align: left;
		color: $ink-1;
		width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.detail {
		text-align: left;
		color: $ink-3;
		max-width: 100%;
		width: 100%;

		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}
}
</style>
