<template>
	<div
		class="justify-center items-center"
		:class="layout"
		v-if="!globalConfig.isOfficial"
	>
		<div
			class="row install_btn_bg"
			@click.stop
			:style="{
				'--textColor': uiState.textColor,
				'--backgroundColor': uiState.backgroundColor,
				'--border': uiState.border,
				'--width': larger ? '108px' : '88px',
				'--statusWidth': larger
					? showMenu
						? 'calc(100% - 25px)'
						: '100%'
					: showMenu
					? 'calc(100% - 21px)'
					: '100%'
			}"
		>
			<q-btn
				:loading="uiState.isLoading"
				:class="larger ? 'application_install_larger' : 'application_install'"
				:style="{
					'--radius': showMenu ? '0' : larger ? '8px' : '4px'
				}"
				@click="debouncedOnclick"
				:disabled="uiState.isDisabled"
				dense
				flat
				no-caps
			>
				<q-tooltip v-if="getErrorTextByState(item.status.state)">
					{{ t(getErrorTextByState(item.status.state)) }}
				</q-tooltip>
				<div>{{ t(uiState.statusText) }}</div>

				<template v-slot:loading>
					<div
						style="width: 100%; height: 100%"
						class="row justify-center items-center"
						v-if="showAppStatus(item.status)"
					>
						{{ t(uiState.statusText) }}
					</div>
					<progress-button
						ref="progressBar"
						v-if="showDownloadProgress(item.status)"
						:progress="item.status.progress"
						:covered-text-color="white"
						:default-text-color="blueDefault"
						:progress-bar-color="blueDefault"
					/>
					<!--				<div v-if="isAppLoading(item.state)">-->
					<!--					<q-img-->
					<!--						class="pending-image"-->
					<!--						:src="getRequireImage('pending_loading.png')"-->
					<!--					/>-->
					<!--				</div>-->
				</template>
			</q-btn>
			<div
				v-if="showMenu"
				class="install_btn_separator_bg items-center"
				:style="{
					background: uiState.backgroundColor,
					height: larger ? '32px' : '24px'
				}"
			>
				<div
					class="install_btn_separator"
					:style="{
						height: larger ? '16px' : '12px',
						marginTop: larger ? '8px' : '6px'
					}"
				/>
			</div>
			<q-btn-dropdown
				v-if="showMore && !deviceStore.isMobile"
				:dropdown-icon="
					uiState.isErrorStatus
						? 'img:/market/market_arrow_error.svg'
						: uiState.isStopStatus
						? 'img:/market/market_arrow_stop.svg'
						: 'img:/market/market_arrow.svg'
				"
				:size="larger ? '12px' : '9px'"
				:class="
					larger
						? 'application_install_larger_more'
						: 'application_install_more'
				"
				content-class="dropdown-menu"
				flat
				dense
				:menu-offset="[0, 4]"
			>
				<div class="column text-body3">
					<div
						v-if="showRetry"
						class="dropdown-menu-item q-mt-xs"
						v-close-popup
						@click="onRetry"
					>
						{{ t('app.retry') }}
					</div>
					<div
						v-if="showResume"
						class="dropdown-menu-item q-mt-xs"
						v-close-popup
						@click="onResume"
					>
						{{ t('app.resume') }}
					</div>
					<div
						v-if="showStop"
						class="dropdown-menu-item q-mt-xs"
						v-close-popup
						@click="onStop"
					>
						{{ t('app.stop') }}
					</div>
					<div
						v-if="showOpenInUpgrade"
						class="dropdown-menu-item q-mt-xs"
						v-close-popup
						@click="onUpdateOpen"
					>
						{{ t('app.open') }}
					</div>
					<div
						v-if="showClone"
						class="dropdown-menu-item q-mt-xs"
						v-close-popup
						@click="onClone"
					>
						{{ t('app.clone') }}
					</div>
					<div
						v-if="showUninstall"
						class="dropdown-menu-item"
						v-close-popup
						@click="onUninstall"
					>
						{{ t('app.uninstall') }}
					</div>
					<div
						v-if="showRemoveLocal"
						class="dropdown-menu-item"
						v-close-popup
						@click="onRemoveLocal"
					>
						{{ t('app.remove') }}
					</div>
				</div>
			</q-btn-dropdown>

			<q-btn
				v-if="showMore && deviceStore.isMobile"
				:icon="
					uiState.isErrorStatus
						? 'img:/market/market_arrow_error.svg'
						: uiState.isStopStatus
						? 'img:/market/market_arrow_stop.svg'
						: 'img:/market/market_arrow.svg'
				"
				:size="larger ? '12px' : '9px'"
				:class="
					larger
						? 'application_install_larger_more'
						: 'application_install_more'
				"
				flat
				dense
				@click.stop="onDropClick"
			/>

			<q-btn
				v-if="showCancelBtn"
				:icon="
					item.status.state === APP_STATUS.DOWNLOAD.DEFAULT
						? 'img:/market/market_download_close.svg'
						: 'img:/market/market_normal_close.svg'
				"
				:size="larger ? '12px' : '9px'"
				:class="
					larger
						? 'application_install_larger_more'
						: 'application_install_more'
				"
				flat
				dense
				@click.stop="onCancelInstall"
			/>
		</div>
		<div
			class="text-overline text-center text-ink-3"
			:class="layout.includes('row') ? 'q-ml-sm' : 'q-mt-xs'"
			v-if="showImageSize"
		>
			{{ calcImageSize }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import InstallButtonOperationDialog from 'src/components/appcard/InstallButtonOperationDialog.vue';
import ProgressButton from '../../components/base/ProgressButton.vue';
import SimpleWaiter from 'src/utils/simpleWaiter';
import { getSuitableUnit, getValueByUnit } from 'src/utils/settings/monitoring';
import useAppAction from 'src/components/appcard/useAppAction';
import { useDeviceStore } from 'src/stores/settings/device';
import { notifyFailed } from 'src/utils/settings/btNotify';
import { usePaymentStore } from 'src/stores/market/payment';
import { useCenterStore } from 'src/stores/market/center';
import { AppService } from 'src/stores/market/appService';
import { useUserStore } from 'src/stores/market/user';
import globalConfig from '../../api/market/config';
import { busOff, busOn } from 'src/utils/bus';
import { useColor } from '@bytetrade/ui';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import debounce from 'lodash/debounce';

import {
	computed,
	onBeforeUnmount,
	onMounted,
	PropType,
	reactive,
	ref,
	watch
} from 'vue';
import {
	canOpen,
	canUpgrade,
	getBaseTextByState,
	getErrorTextByState,
	getPaymentTextByState,
	isCanceling,
	isDoing,
	isFailed,
	showAppStatus,
	showDownloadProgress,
	uninstalledApp
} from 'src/constant/config';
import {
	APP_PAYMENT_TYPE,
	APP_STATUS,
	AppStatusLatest,
	LOCAL_STATUS,
	PAYMENT_STATUS
} from 'src/constant/constants';

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
	},
	layout: {
		type: String,
		default: 'row'
	},
	imageSize: {
		type: Boolean,
		default: true
	}
});
const { color: blueDefault } = useColor('blue-default');
const { color: blueAlpha } = useColor('blue-alpha');
const { color: orangeSoft } = useColor('orange-soft');
const { color: orangeDefault } = useColor('orange-default');
const { color: redSoft } = useColor('red-soft');
const { color: negative } = useColor('negative');
const { color: grey } = useColor('background-3');
const { color: white } = useColor('ink-on-brand');
const emit = defineEmits(['onError']);
const simpleWaiter = new SimpleWaiter();
const appPaymentType = ref<APP_PAYMENT_TYPE>(APP_PAYMENT_TYPE.APP_TYPE_UNKNOWN);
const deviceStore = useDeviceStore();
const centerStore = useCenterStore();
const paymentStore = usePaymentStore();
const userStore = useUserStore();
const { t } = useI18n();
const $q = useQuasar();

const {
	showRemoveLocal,
	showCancelBtn,
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
	openApp,
	showClone,
	onClone,
	showMore,
	showMenu,
	onCancelInstall
} = useAppAction(props);

type UIStateType =
	| 'uninstalled'
	| 'unpaid'
	| 'installable'
	| 'error'
	| 'doing'
	| 'canceling'
	| 'installed'
	| 'update'
	| 'stopCompleted'
	| 'running'
	| 'defaultError';

interface UIConfig {
	isDisabled: boolean;
	isLoading: boolean;
	isErrorStatus: boolean;
	isStopStatus: boolean;
	statusText: string;
	textColor: string;
	backgroundColor: string;
	border: string;
}

type StateChecker = () => boolean;

const uiState = reactive<UIConfig>({
	isDisabled: false,
	isLoading: false,
	isErrorStatus: false,
	isStopStatus: false,
	statusText: '',
	textColor: '',
	backgroundColor: '',
	border: ''
});

const currentAppLocalStatus = computed(() => {
	return centerStore.getLocalStatus(props.appName, props.sourceId);
});

const currentStateMeta = ref<{
	type: UIStateType;
	uiConfig: UIConfig | ((params: any) => UIConfig);
	handler?: (params: any) => Promise<void>;
} | null>(null);

const stateParams = ref<{
	configParams: any;
	handlerParams: any;
}>({
	configParams: '',
	handlerParams: ''
});

const stateConfigs: Array<{
	type: UIStateType;
	checker: StateChecker;
	uiConfig: UIConfig | ((params: any) => UIConfig);
	handler?: (params: any) => Promise<void>;
}> = [
	{
		type: 'uninstalled',
		checker: () =>
			uninstalledApp(props.item.status) &&
			(appPaymentType.value === APP_PAYMENT_TYPE.APP_TYPE_UNKNOWN ||
				!currentAppLocalStatus.value ||
				(currentAppLocalStatus.value.status ===
					LOCAL_STATUS.PRE_CHECK_FINISHED &&
					currentAppLocalStatus.value.data.length > 0)),
		uiConfig: {
			statusText: 'app.get',
			textColor: blueDefault.value,
			backgroundColor: grey.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: false
		},
		handler: async () => {
			const appFullInfo = centerStore.getAppFullInfo(
				props.appName,
				props.sourceId
			);
			if (!appFullInfo) {
				notifyFailed('Fetching the full data of the current app, please wait.');
				return;
			}

			const errorGroup = userStore.frontendPreflight(
				appFullInfo.app_info.app_entry
			);
			emit('onError', errorGroup);

			console.log(appFullInfo);

			if (
				appPaymentType.value === APP_PAYMENT_TYPE.APP_TYPE_PAID &&
				errorGroup.length === 0
			) {
				await paymentStore.fetchPaymentInfo(props.appName, props.sourceId);
			} else {
				centerStore.updateLocalStatus(props.appName, props.sourceId, {
					status: LOCAL_STATUS.PRE_CHECK_FINISHED,
					data: errorGroup
				});
			}
		}
	},

	{
		type: 'unpaid',
		checker: () => {
			if (currentAppLocalStatus.value) {
				stateParams.value.configParams = currentAppLocalStatus.value.status;
			}
			return (
				uninstalledApp(props.item.status) &&
				appPaymentType.value === APP_PAYMENT_TYPE.APP_TYPE_PAID &&
				currentAppLocalStatus.value &&
				currentAppLocalStatus.value.status !== PAYMENT_STATUS.PURCHASED &&
				currentAppLocalStatus.value.status !== PAYMENT_STATUS.NOT_EVALUATED
			);
		},
		uiConfig: (params) => ({
			statusText: getPaymentTextByState(params),
			textColor: blueDefault.value,
			backgroundColor: blueAlpha.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: false
		}),
		handler: async () => {
			console.log('starting payment');
			await paymentStore.queryPaymentInfo(props.appName, props.sourceId, t, $q);
		}
	},

	{
		type: 'installable',
		checker: () => {
			if (!uninstalledApp(props.item.status)) {
				return false;
			}
			if (appPaymentType.value === APP_PAYMENT_TYPE.APP_TYPE_PAID) {
				return currentAppLocalStatus.value.status === PAYMENT_STATUS.PURCHASED;
			}
			return true;
		},
		uiConfig: {
			statusText: 'app.install',
			textColor: blueDefault.value,
			backgroundColor: blueAlpha.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: false
		},
		handler: async () => {
			AppService.installApp(
				props.item.status,
				{
					app_name: props.appName,
					source: props.sourceId,
					version: props.version
				},
				$q
			);
		}
	},

	{
		type: 'error',
		checker: () => {
			return (
				isFailed(props.item.status) ||
				(currentAppLocalStatus.value &&
					currentAppLocalStatus.value.status === PAYMENT_STATUS.NOT_EVALUATED)
			);
		},
		uiConfig: () => ({
			statusText: t('app.failed'),
			textColor: negative.value,
			backgroundColor: redSoft.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: true,
			isErrorStatus: true,
			isStopStatus: false
		})
	},

	{
		type: 'doing',
		checker: () => {
			stateParams.value.configParams = props.item.status.state;
			return isDoing(props.item.status);
		},
		uiConfig: (params) => ({
			statusText: getBaseTextByState(params),
			textColor: white.value,
			backgroundColor:
				props.item.status.state === APP_STATUS.DOWNLOAD.DEFAULT
					? blueAlpha.value
					: blueDefault.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: true,
			isErrorStatus: false,
			isStopStatus: false
		})
	},

	{
		type: 'canceling',
		checker: () => isCanceling(props.item.status),
		uiConfig: {
			statusText: 'app.canceling',
			textColor: white.value,
			backgroundColor: blueDefault.value,
			border: '1px solid transparent',
			isDisabled: true,
			isLoading: true,
			isErrorStatus: false,
			isStopStatus: false
		}
	},

	{
		type: 'installed',
		checker: () => props.item.status.state === APP_STATUS.MODEL.INSTALLED,
		uiConfig: {
			statusText: 'app.installed',
			textColor: blueDefault.value,
			backgroundColor: blueAlpha.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: false
		}
	},

	{
		type: 'update',
		checker: () => {
			const { state } = props.item.status;
			return (
				(state === APP_STATUS.STOP.COMPLETED || state === APP_STATUS.RUNNING) &&
				canUpgrade(props.item, props.appName, props.sourceId)
			);
		},
		uiConfig: {
			statusText: 'app.update',
			textColor: blueDefault.value,
			backgroundColor: blueAlpha.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: false
		},
		handler: async () => {
			AppService.upgradeApp(props.item.status, {
				app_name: props.appName,
				source: props.sourceId,
				version: props.version
			});
		}
	},

	{
		type: 'stopCompleted',
		checker: () =>
			props.item.status.state === APP_STATUS.STOP.COMPLETED &&
			!canUpgrade(props.item, props.appName, props.sourceId),
		uiConfig: {
			statusText: 'app.stopped',
			textColor: orangeDefault.value,
			backgroundColor: orangeSoft.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: true
		}
	},

	{
		type: 'running',
		checker: () => {
			stateParams.value.configParams = props.item.status;
			return (
				props.item.status.state === APP_STATUS.RUNNING &&
				!canUpgrade(props.item, props.appName, props.sourceId)
			);
		},
		uiConfig: (status: any) => ({
			statusText: canOpen(status) ? 'app.open' : 'app.running',
			textColor: blueDefault.value,
			backgroundColor: blueAlpha.value,
			border: '1px solid transparent',
			isDisabled: false,
			isLoading: false,
			isErrorStatus: false,
			isStopStatus: false
		}),
		handler: async () => {
			if (canOpen(props.item.status)) {
				openApp();
			}
		}
	},

	{
		type: 'defaultError',
		checker: () => false,
		uiConfig: (state: string) => ({
			statusText: state,
			textColor: negative.value,
			backgroundColor: redSoft.value,
			border: '1px solid transparent',
			isDisabled: true,
			isLoading: false,
			isErrorStatus: true,
			isStopStatus: false
		})
	}
];

const onDropClick = () => {
	if (!deviceStore.isMobile) {
		return;
	}
	$q.dialog({
		component: InstallButtonOperationDialog,
		componentProps: {
			item: props.item,
			appName: props.appName,
			version: props.version,
			larger: props.larger,
			sourceId: props.sourceId,
			manager: props.manager
		}
	}).onOk(() => {});
};

watch(
	() => props.item,
	() => {
		if (props.item) {
			updateUIAndMeta();
		}
	},
	{
		deep: true,
		immediate: true
	}
);

const showImageSize = computed(() => {
	return (
		props.imageSize &&
		(currentStateMeta.value.type === 'installable' ||
			currentStateMeta.value.type === 'update')
	);
});

const calcImageSize = computed(() => {
	let needDownloadSize = 0;
	const fullLatest = centerStore.getAppFullInfo(props.appName, props.sourceId);
	if (fullLatest?.app_info?.image_analysis?.images) {
		Object.values(fullLatest?.app_info?.image_analysis?.images).forEach(
			(item: any) => {
				needDownloadSize += item.total_size - item.downloaded_size;
			}
		);
	}
	return (
		getValueByUnit(
			needDownloadSize,
			getSuitableUnit(needDownloadSize, 'disk')
		) + getSuitableUnit(needDownloadSize, 'disk')
	);
});

function resetUIState() {
	Object.assign(uiState, {
		isDisabled: false,
		isLoading: false,
		isErrorStatus: false,
		isStopStatus: false,
		statusText: '',
		textColor: '',
		backgroundColor: '',
		border: ''
	});
	currentStateMeta.value = null;
	stateParams.value = {
		configParams: '',
		handlerParams: ''
	};
}

async function updateUIAndMeta() {
	resetUIState();

	for (const config of stateConfigs) {
		const result = config.checker();
		if (result) {
			currentStateMeta.value = {
				type: config.type,
				uiConfig: config.uiConfig,
				handler: config.handler
			};

			const uiConfig =
				typeof config.uiConfig === 'function'
					? config.uiConfig(stateParams.value.configParams)
					: config.uiConfig;
			Object.assign(uiState, uiConfig);
			// console.log(config);
			// console.log(currentAppLocalStatus.value);
			return;
		}
	}

	const defaultConfig = stateConfigs.find(
		(c) => c.type === 'defaultError' || c.type === 'error'
	);
	if (defaultConfig) {
		currentStateMeta.value = {
			type: defaultConfig.type,
			uiConfig: defaultConfig.uiConfig,
			handler: defaultConfig.handler
		};
		const uiConfig =
			typeof defaultConfig.uiConfig === 'function'
				? defaultConfig.uiConfig(props.item.status?.state || 'unknown_error')
				: defaultConfig.uiConfig;
		Object.assign(uiState, uiConfig);
	}
}

const updateUiByLocal = (data: any) => {
	if (data.appId !== props.appName || data.sourceId !== props.sourceId) {
		return;
	}
	updateUIAndMeta();
};

onMounted(() => {
	busOn('local_state_update', updateUiByLocal);
	simpleWaiter.waitForCondition(
		() => {
			const appFullInfo = centerStore.getAppFullInfo(
				props.appName,
				props.sourceId
			);

			return !!appFullInfo;
		},
		async () => {
			const appFullInfo = centerStore.getAppFullInfo(
				props.appName,
				props.sourceId
			);
			appPaymentType.value = appFullInfo?.app_info?.price
				? APP_PAYMENT_TYPE.APP_TYPE_PAID
				: APP_PAYMENT_TYPE.APP_TYPE_FREE;
			await updateUIAndMeta();
		}
	);
});

onBeforeUnmount(() => {
	busOff('local_state_update', updateUiByLocal);
});

function validateBeforeClick(): boolean {
	if (!props.item || !props.item.status) {
		console.log(`install button item does not exist`);
		return false;
	}
	return true;
}

const onClick = async () => {
	console.log('onClick');
	if (!validateBeforeClick()) return;

	if (currentStateMeta.value?.handler) {
		try {
			await currentStateMeta.value.handler(stateParams.value.handlerParams);
			await updateUIAndMeta();
		} catch (error) {
			console.error('onClick error：', error);
		}
	}
};

const debouncedOnclick = debounce(onClick, 500, {
	maxWait: 2000
});
</script>

<style scoped lang="scss">
.pending-image {
	width: 12px;
	height: 12px;
	animation: animate 1.2s linear infinite;
	-webkit-animation: animate 1.2s linear infinite;
}

.install_btn_bg {
	width: var(--width);
	min-width: var(--width);
	max-width: var(--width);
	border-radius: 4px;
	padding: 0;

	.install_btn_separator_bg {
		height: 100%;
		padding-left: 1px;
		padding-right: 1px;

		.install_btn_separator {
			width: 1px;
			background: $btn-stroke;
		}
	}

	.application_install {
		box-sizing: border-box;
		width: var(--statusWidth);
		min-width: var(--statusWidth);
		max-width: var(--statusWidth);
		color: var(--textColor);
		background: var(--backgroundColor);
		border-radius: 4px var(--radius, 0) var(--radius, 0) 4px !important;
		height: 24px;
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		font-family: Roboto;
		font-size: 12px;
		font-weight: 500;
		line-height: 16px;
		letter-spacing: 0em;
		text-align: center;
		border: var(--border);
	}

	.application_install_more {
		width: 18px;
		color: var(--textColor);
		background: var(--backgroundColor);
		height: 24px;
		border-radius: 0 4px 4px 0 !important;
		gap: 20px;
		text-align: center;
	}

	.application_install_larger {
		box-sizing: border-box;
		width: var(--statusWidth);
		min-width: var(--statusWidth);
		max-width: var(--statusWidth);
		color: var(--textColor);
		background: var(--backgroundColor);
		height: 32px;
		border-radius: 8px var(--radius, 0) var(--radius, 0) 8px !important;
		font-family: Roboto;
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
		font-size: 14px;
		font-weight: 500;
		line-height: 20px;
		letter-spacing: 0;
		text-align: center;
		border: var(--border);
	}

	.application_install_larger_more {
		width: 22px;
		color: var(--textColor);
		background: var(--backgroundColor);
		height: 32px;
		border-radius: 0 8px 8px 0 !important;
		text-align: center;
	}
}

.dropdown-menu-item {
	height: 32px;
	color: $ink-2;
	padding: 8px 12px;

	&:hover {
		background: $background-hover;
		border-radius: 4px;
	}

	&:active {
		background: $background-hover;
		border-radius: 4px;
	}
}
</style>
