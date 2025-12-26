<template>
	<application-item
		:icon="app.icon"
		:title="app.title"
		:status="app.state"
		:hide-status="false"
		:width-separator="false"
		:end-icon="false"
		:margin-top="false"
	>
		<template v-slot>
			<bt-switch
				v-if="(adminStore.isAdmin || !isDemo) && !app.isSysApp"
				truthy-track-color="blue-default"
				:disable="
					app?.state !== APP_STATUS.STOP.COMPLETED &&
					app?.state !== APP_STATUS.RUNNING
				"
				v-model="isRunning"
				@click="toggle"
			/>
		</template>
	</application-item>
</template>

<script lang="ts" setup>
import ApplicationItem from './ApplicationItem.vue';
import { useApplicationStore } from 'src/stores/settings/application';
import { canResumeState, canStopState } from 'src/constant/config';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useAdminStore } from 'src/stores/settings/admin';
import { APP_STATUS } from 'src/constant/constants';
import { computed, PropType, ref } from 'vue';
import { TerminusApp } from '@bytetrade/core';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const $q = useQuasar();
const applicationStore = useApplicationStore();
const adminStore = useAdminStore();
const { t } = useI18n();
const isDemo = computed(() => {
	return !!process.env.DEMO;
});

const props = defineProps({
	app: {
		type: Object as PropType<TerminusApp>,
		required: true
	}
});

const isRunning = computed(
	() => props.app && props.app.state == APP_STATUS.RUNNING
);

async function toggle() {
	console.log(props.app?.state);
	if (!canResumeState(props.app?.state) && !canStopState(props.app?.state)) {
		notifyFailed(t('app_operation_not_allowed'));
		return;
	}
	$q.loading.show();

	try {
		if (canResumeState(props.app?.state)) {
			await applicationStore.resume(props.app.name);
		} else if (canStopState(props.app?.state)) {
			await applicationStore.suspend(props.app.name);
		}
		await applicationStore.status(props.app.name);
	} catch (e) {
		/* empty */
	} finally {
		$q.loading.hide();
	}
}
</script>

<style scoped lang="scss">
.application-detail-item {
	// height: 32px;
	// min-height: 32px;
	padding: 0;
	// background-color: red;
	height: 56px;

	.application-logo {
		width: 32px;
		height: 32px;
		border-radius: 8px;
	}

	.application-name {
		text-transform: capitalize;
		color: $ink-1;
		margin-left: 8px;
	}

	.application-status {
		text-align: right;
		color: $ink-2;
		text-transform: capitalize;
		margin-left: 4px;
	}
}
</style>
