<template>
	<page-title-component :show-back="true" :title="t('logs')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list>
			<bt-form-item
				:title="t('get_logs')"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="collected"
			>
				<div class="upgradeNow" v-if="!firstLoad">
					{{ t('checking_system_status') }}
				</div>
				<div
					class="upgradeNow"
					v-if="firstLoad && collectingStatus === ProcessingState.InProgress"
				>
					<span class="loader" />
					{{ t('collecting') }}
				</div>
				<div
					class="upgradeNow"
					v-if="
						firstLoad &&
						(collectingStatus === ProcessingState.Init || collected)
					"
					@click="handleCollect"
				>
					{{ t('collect') }}
				</div>
			</bt-form-item>

			<bt-form-item
				:title="
					collectingStatus === ProcessingState.Completed
						? t('log_file_storage')
						: t('collect_failed_reason')
				"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="false"
				v-if="collected"
			>
				<div
					v-if="collectingStatus === ProcessingState.Completed"
					class="row justify-end items-center"
				>
					<div class="text-body2 q-mr-lg">/Home/pod_logs</div>
					<div class="upgradeNow" @click="openLogFile">
						{{ t('base.open') }}
					</div>
				</div>

				<div
					v-if="
						collectingStatus === ProcessingState.Failed && collectFailedReason
					"
					class="text-body2 q-mr-lg text-negative"
					style="text-align: end; word-break: break-all"
				>
					{{ collectFailedReason }}
				</div>
			</bt-form-item>
		</bt-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { useTerminusDStore } from 'src/stores/settings/terminusd';
import { ProcessingState } from 'src/constant/global';
import BtList from 'src/components/settings/base/BtList.vue';
import * as process from 'process';

const terminusDStore = useTerminusDStore();
const { t } = useI18n();

const collectFailedReason = ref('');
const collectingStatus = ref<ProcessingState>(ProcessingState.Init);
const firstLoad = ref(false);
const timer = ref();
const collected = computed(() => {
	return (
		collectingStatus.value === ProcessingState.Completed ||
		collectingStatus.value === ProcessingState.Failed
	);
});

const handleCollect = async () => {
	terminusDStore
		.collect_logs()
		.then(() => {
			collectingStatus.value = ProcessingState.InProgress;
			startQueryTimer();
		})
		.catch(() => {
			notifyFailed(t('failed'));
		});
};

const openLogFile = () => {
	const hostname = window.location.hostname.replace(
		process.env.DEV ? 'test' : 'settings',
		'files'
	);
	let path =
		window.location.protocol + '//' + hostname + '/Files/Home/pod_logs/';
	window.open(path);
};

const startQueryTimer = async () => {
	await queryLogStatus();
	firstLoad.value = true;
	if (collectingStatus.value === ProcessingState.InProgress) {
		timer.value = setInterval(async () => {
			if (collected.value) {
				window.clearInterval(timer.value);
				return;
			}
			await queryLogStatus();
		}, 3000);
	}
};

const queryLogStatus = async () => {
	try {
		await terminusDStore.system_status();
		collectingStatus.value = terminusDStore.olaresInfo.collectingLogsState;
		collectFailedReason.value = terminusDStore.olaresInfo.collectingLogsError;
	} catch (error) {
		console.log(error);
	}
};

onMounted(() => {
	startQueryTimer();
});

onUnmounted(() => {
	window.clearInterval(timer.value);
});
</script>

<style scoped lang="scss">
.list_section {
	min-height: 56px;
}

.upgradeNow {
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
