import { notifyFailed } from 'src/utils/settings/btNotify';
import { useCenterStore } from 'src/stores/market/center';
import { AppService } from 'src/stores/market/appService';
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
	isUninstallable,
	uninstalledApp
} from 'src/constant/config';
import {
	APP_STATUS,
	ENTRANCE_STATUS,
	MARKET_SOURCE_OFFICIAL
} from 'src/constant/constants';

export default function useAppAction(props) {
	const { t, locale } = useI18n();
	const { color: blueDefault } = useColor('blue-default');
	const { color: white } = useColor('ink-on-brand');
	const centerStore = useCenterStore();
	const $q = useQuasar();

	const showMore = computed(() => {
		return (
			props.manager &&
			props.item &&
			props.item.status &&
			(showResume.value ||
				showUninstall.value ||
				showOpenInUpgrade.value ||
				showStop.value ||
				showRetry.value ||
				showRemoveLocal.value ||
				showClone.value)
		);
	});

	const showCancelBtn = computed(() => {
		return canInstallingCancel(props.item.status);
	});

	const supportClone = computed(() => {
		const aggregation = centerStore.getAppAggregationInfo(
			props.appName,
			props.sourceId
		);
		return aggregation?.app_full_info?.app_info?.app_entry?.options
			?.allowMultipleInstall;
	});

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

	const showUninstall = computed(() => {
		return isUninstallable(props.item.status);
	});

	const showOpenInUpgrade = computed(() => {
		return canUpgrade(props.item, props.appName, props.sourceId);
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

	function onClone() {
		if (!props.item || !props.item.status) {
			return;
		}
		AppService.cloneApp(
			props.item.status,
			{
				app_name: props.appName,
				source: props.sourceId
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
				version: props.version
			});
		}
	}

	async function onRemoveLocal() {
		if (!props.item || !props.item.status) {
			return;
		}
		if (uninstalledApp(props.item.status)) {
			//has clone app return
			if (supportClone.value) {
				const cloneAppList: string[] = [];
				centerStore.appStatusMap.forEach((statusLatest, combinedId) => {
					if (
						statusLatest &&
						isCloneApp(statusLatest.status) &&
						statusLatest.status.rawAppName === props.appName
					) {
						cloneAppList.push(combinedId);
					}
				});

				if (cloneAppList.length > 0) {
					notifyFailed(
						'Removal of the installation package is prohibited as there are cloned versions of the current application.'
					);
					return;
				}
			}
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
					version: props.version
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
						version: props.version
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
					version: props.version
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
		openApp,
		onCancelInstall
	};
}
