import { AppService } from 'src/services/market/appOperation/appService';
import { useAppStore } from 'src/stores/market/appStore';
import { BtDialog, useColor } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { computed } from 'vue';
import {
	canInstallingCancel,
	canOpen,
	canResume,
	canStop,
	canUpgrade,
	isCloneApp,
	isFailed,
	isMultiInstanceApp,
	isTemplateApp,
	isTemplateInstance,
	isUninstallable,
	uninstalledApp
} from 'src/constant/config';
import {
	APP_STATUS,
	ENTRANCE_STATUS,
	MARKET_SOURCE_OFFICIAL
} from 'src/constant/constants';

export default function useAppAction(props: any) {
	const { t } = useI18n();
	const { color: blueDefault } = useColor('blue-default');
	const { color: white } = useColor('ink-on-brand');
	const appStore = useAppStore();
	const $q = useQuasar();

	const showMore = computed(() => {
		return (
			props.manager &&
			props.item &&
			props.item.status &&
			(showResume.value ||
				showUninstall.value ||
				showOpenInUpgrade.value ||
				showUpdateInStop.value ||
				showStop.value ||
				showRetry.value ||
				showRemoveLocal.value ||
				showClone.value)
		);
	});

	const showCancelBtn = computed(() => {
		return canInstallingCancel(props.item.status);
	});

	const appSimpleInfo = computed(
		() =>
			appStore.getAppAggregationInfo(props.appName, props.sourceId)
				?.app_simple_latest?.app_simple_info
	);

	const supportClone = computed(() => isMultiInstanceApp(appSimpleInfo.value));

	// Is the current app a template body (no installable body, Create instances).
	// A clone instance shares the template chart's simpleInfo (so templateOnly is
	// also true on it); exclude clones so instances are treated as regular apps.
	const isTemplate = computed(
		() => isTemplateApp(appSimpleInfo.value) && !isCloneApp(props.item?.status)
	);

	// Is the current status a template instance (a clone whose own simpleInfo is
	// a template). Read from the instance's own simple — never look up the body.
	const isTemplateInstanceApp = computed(() =>
		isTemplateInstance(props.item?.status, appSimpleInfo.value)
	);

	// Number of instances created from this template body.
	const instancesCount = computed(
		() => appStore.getCloneApp(props.appName, props.sourceId).length
	);

	const showClone = computed(() => {
		return (
			supportClone.value &&
			props.item.status?.rawAppName === props.item.status?.name &&
			props.item.status?.state === APP_STATUS.RUNNING
		);
	});

	const showRemoveLocal = computed(() => {
		return (
			uninstalledApp(props.item.status) &&
			props.sourceId === MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD &&
			props.item.status?.rawAppName === props.item.status?.name
		);
	});

	const showTemplateRemove = computed(
		() => props.manager && showRemoveLocal.value
	);

	const showUninstall = computed(() => {
		return isUninstallable(props.item.status);
	});

	const showOpenInUpgrade = computed(() => {
		return (
			props.item.status.state === APP_STATUS.RUNNING &&
			canUpgrade(props.item, props.appName, props.sourceId)
		);
	});

	const showUpdateInStop = computed(() => {
		return (
			props.item.status.state === APP_STATUS.STOP.COMPLETED &&
			canUpgrade(props.item, props.appName, props.sourceId)
		);
	});

	const showStop = computed(() => {
		return canStop(props.item.status);
	});

	const showResume = computed(() => {
		return canResume(props.item.status);
	});

	const showRetry = computed(() => {
		return (
			isFailed(props.item.status) &&
			props.item.status.state !== APP_STATUS.ENV.APPLY_FAILED &&
			props.item.status.state !== APP_STATUS.ENV.CANCEL_FAILED
		);
	});

	// Reused for both "Clone" (multi-instance body) and "Create" (template body).
	// For a template body the body itself is never installed, so app_name is the
	// template chart name.
	function onClone() {
		if (!props.item || !props.item.status) {
			return;
		}
		AppService.cloneApp(
			props.item.status,
			{
				app_name: props.appName,
				source: props.sourceId,
				// Template body has no installable body; flag template instance creation
				// so the backend and the clone dialog title can branch accordingly.
				...(isTemplate.value ? { templateClone: true } : {})
			},
			$q
		);
	}

	function onStop() {
		if (!props.item || !props.item.status) {
			return;
		}
		AppService.stopApp(
			props.item.status,
			{
				app_name: props.appName,
				source: props.sourceId
			},
			$q
		);
	}

	function onResume() {
		if (!props.item || !props.item.status) {
			return;
		}
		AppService.resumeApp(
			props.item.status,
			{
				app_name: props.appName,
				source: props.sourceId
			},
			$q
		);
	}

	function onUpdate() {
		if (!props.item || !props.item.status) {
			return;
		}
		AppService.upgradeApp(props.item.status, {
			app_name: props.appName,
			source: props.sourceId,
			version: props.version
		});
	}

	function onUpdateOpen() {
		if (!props.item || !props.item.status) {
			return;
		}
		switch (props.item.status?.state) {
			case APP_STATUS.RUNNING:
				openApp();
				break;
		}
	}

	const openApp = () => {
		if (!props.item || !props.item.status) {
			return;
		}

		const entrance = canOpen(props.item.status);
		if (entrance) {
			switch (entrance.state) {
				case ENTRANCE_STATUS.RUNNING:
					AppService.openApp(entrance);
					break;
				case ENTRANCE_STATUS.STOPPED:
					onSuspendTips(true);
					break;
				case ENTRANCE_STATUS.NOT_READY:
					onSuspendTips(false);
					break;
			}
		}
	};

	async function onSuspendTips(isStop: boolean) {
		if (!props.item || !props.item.status) {
			return;
		}
		BtDialog.show({
			title: isStop ? t('Entrance paused') : t('Getting your app ready'),
			message: isStop
				? t(
						'Entrance to this application is currently paused. Please try resuming the app to re-start it.'
				  )
				: t('This can sometimes take a few moments. Thanks for your patience.'),
			okStyle: {
				background: blueDefault.value,
				color: white.value
			},
			okText: t('base.ok'),
			cancel: false
		})
			.then((res) => {
				if (res) {
					console.log('click ok');
				} else {
					console.log('click cancel');
				}
			})
			.catch((err) => {
				console.log('click error', err);
			});
	}

	const showMenu = computed(() => {
		return showMore.value || showCancelBtn.value;
	});

	async function onCancelInstall() {
		if (canInstallingCancel(props.item.status)) {
			console.log(props.item.status?.name);
			console.log('cancel installing');
			await AppService.cancelInstallingApp(props.item.status, {
				app_name: props.appName,
				source: props.sourceId,
				version: props.item?.version
			});
		}
	}

	async function onRemoveLocal() {
		if (!props.item || !props.item.status) {
			return;
		}
		if (uninstalledApp(props.item.status)) {
			BtDialog.show({
				title: t('app.remove'),
				message: t(
					'Are you sure you want to delete this app chart from Local Sources?'
				),
				okStyle: {
					background: blueDefault.value,
					color: white.value
				},
				okText: t('base.confirm'),
				cancel: true
			})
				.then(async (res) => {
					if (res) {
						console.log('click ok');
						await AppService.removeApp(props.item.status, {
							app_name: props.appName,
							source: MARKET_SOURCE_OFFICIAL.LOCAL.UPLOAD,
							version: props.version
						});
					} else {
						console.log('click cancel');
					}
				})
				.catch((err) => {
					console.log('click error', err);
				});
		}
	}

	const onRetry = async () => {
		if (!props.item || !props.item.status) {
			return;
		}
		switch (props.item.status?.state) {
			case APP_STATUS.PENDING.CANCEL_FAILED:
			case APP_STATUS.DOWNLOAD.CANCEL_FAILED:
			case APP_STATUS.INSTALL.CANCEL_FAILED:
				await AppService.cancelInstallingApp(props.item.status, {
					app_name: props.appName,
					source: props.sourceId,
					version: props.item?.version
				});
				break;
			case APP_STATUS.DOWNLOAD.FAILED:
			case APP_STATUS.INSTALL.FAILED:
				await AppService.installApp(
					props.item.status,
					{
						app_name: props.appName,
						source: props.sourceId,
						version: props.version
					},
					$q
				);
				break;
			case APP_STATUS.UNINSTALL.FAILED:
				await AppService.uninstallApp(
					props.item.status,
					{
						app_name: props.appName,
						source: props.sourceId,
						version: props.item?.version,
						deleteData: false
					},
					$q
				);
				break;
			case APP_STATUS.UPGRADE.FAILED:
				await AppService.upgradeApp(props.item.status, {
					app_name: props.appName,
					source: props.sourceId,
					version: props.version
				});
				break;
			case APP_STATUS.RESUME.FAILED:
				await AppService.resumeApp(
					props.item.status,
					{
						app_name: props.appName,
						source: props.sourceId
					},
					$q
				);
				break;
			case APP_STATUS.STOP.FAILED:
				await AppService.stopApp(
					props.item.status,
					{
						app_name: props.appName,
						source: props.sourceId
					},
					$q
				);
				break;
		}
	};

	async function onUninstall() {
		if (!props.item || !props.item.status) {
			return;
		}
		if (isUninstallable(props.item.status)) {
			await AppService.uninstallApp(
				props.item.status,
				{
					app_name: props.appName,
					source: props.sourceId,
					version: props.item?.version,
					deleteData: false
				},
				$q
			);
		}
	}

	return {
		showCancelBtn,
		showMenu,
		showMore,
		showClone,
		onClone,
		isTemplate,
		isTemplateInstanceApp,
		instancesCount,
		showRemoveLocal,
		showTemplateRemove,
		onRemoveLocal,
		showUninstall,
		onUninstall,
		showOpenInUpgrade,
		onUpdateOpen,
		showUpdateInStop,
		onUpdate,
		showRetry,
		onRetry,
		showStop,
		onStop,
		showResume,
		onResume,
		openApp,
		onCancelInstall
	};
}
