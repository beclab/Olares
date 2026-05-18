<template>
	<application-item
		:icon="app.icon"
		:title="app.title"
		:status="app.state"
		:app-name="app.name"
		:raw-app-name="app.rawAppName"
		:cs-app="app.isClusterScoped"
		:hide-status="false"
		:width-separator="false"
		:end-icon="false"
		:margin-top="false"
	>
		<template v-slot>
			<div v-if="(adminStore.isAdmin || !isDemo) && !app.isSysApp">
				<div
					v-if="
						app?.state !== APP_STATUS.STOP.COMPLETED &&
						app?.state !== APP_STATUS.RUNNING
					"
					class="btn-operate"
					disabled
				>
					<span class="loader" />
				</div>
				<div v-else-if="isRunning" class="btn-operate" @click="handleStop">
					{{ t('app.stop') }}
				</div>
				<div class="btn-operate" v-else @click="handleResume">
					{{ t('app.resume') }}
				</div>
			</div>
		</template>
	</application-item>
</template>

<script lang="ts" setup>
import ApplicationItem from './ApplicationItem.vue';
import { useApplicationStore } from 'src/stores/settings/application';
import { canResumeState, canStopState } from 'src/constant/config';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useAdminStore } from 'src/stores/settings/admin';
import { useAppStore } from 'src/stores/market/appStore';
import { AppService } from 'src/stores/market/appService';
import { APP_STATUS } from 'src/constant/constants';
import { computed, PropType } from 'vue';
import { TerminusApp } from '@bytetrade/core';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const $q = useQuasar();
const applicationStore = useApplicationStore();
const adminStore = useAdminStore();
const appStore = useAppStore();
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

async function handleStop() {
	console.log(props.app?.state);
	if (!canStopState(props.app?.state)) {
		notifyFailed(t('app_operation_not_allowed'));
		return;
	}

	try {
		if (canStopState(props.app?.state)) {
			const marketApp = appStore.findAppByName(props.app?.name);
			if (marketApp) {
				await AppService.stopApp(
					marketApp.status,
					{
						app_name: marketApp.appId,
						source: marketApp.sourceId
					},
					$q
				);
			} else {
				notifyFailed('Failed to retrieve app information');
			}
		}
		await applicationStore.status(props.app.name);
	} catch (e) {
		/* empty */
	} finally {
		$q.loading.hide();
	}
}

async function handleResume() {
	console.log(props.app?.state);
	if (!canResumeState(props.app?.state)) {
		notifyFailed(t('app_operation_not_allowed'));
		return;
	}
	$q.loading.show();

	try {
		if (canResumeState(props.app?.state)) {
			const marketApp = appStore.findAppByName(props.app?.name);
			if (marketApp) {
				await AppService.resumeApp(
					marketApp.status,
					{
						app_name: marketApp.appId,
						source: marketApp.sourceId
					},
					$q
				);
			} else {
				notifyFailed('Failed to retrieve app information');
			}
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

.btn-operate {
	border: 1px solid $btn-stroke;
	padding: 8px 12px;
	border-radius: 8px;
	cursor: pointer;
	// background-color: $background-3;

	&:hover {
		background: $background-5;
	}
}

.loader {
	width: 1rem;
	height: 1rem;
	color: inherit;
	vertical-align: middle;
	border: 0.2em solid transparent;
	border-top-color: currentcolor;
	border-bottom-color: currentcolor;
	border-radius: 50%;
	position: relative;
	animation: 1s loader-30 linear infinite;
	display: inline-block;

	&:before,
	&:after {
		content: '';
		display: block;
		width: 0;
		height: 0;
		position: absolute;
		border: 0.2em solid transparent;
		border-bottom-color: currentcolor;
	}

	&:before {
		transform: rotate(135deg);
		right: -0.2em;
		top: -0.05em;
	}

	&:after {
		transform: rotate(-45deg);
		left: -0.2em;
		bottom: -0.05em;
	}
}

@keyframes loader-30 {
	0% {
		transform: rotate(0deg);
	}

	100% {
		transform: rotate(360deg);
	}
}
</style>
