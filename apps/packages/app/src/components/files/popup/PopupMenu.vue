<template>
	<q-menu
		@update:model-value="showPopupProxy"
		class="popup-menu bg-background-2"
	>
		<q-list dense padding>
			<q-item
				class="row items-center justify-start text-ink-2 popup-item"
				clickable
				v-close-popup
				v-for="item in menuList"
				:key="item.action"
				@click="handleEvent(item.action, $event)"
			>
				<q-icon :name="item.icon" size="20px" class="q-mr-sm" />
				<q-item-section class="menuName"> {{ t(item.name) }}</q-item-section>
			</q-item>
		</q-list>
	</q-menu>
</template>
<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { seahub } from '../../../api';
import { popupMenu, OPERATE_ACTION, SYNC_STATE } from '../../../utils/contact';
import { useDataStore } from '../../../stores/data';
import { useMenuStore } from '../../../stores/files-menu';
import { useFilesStore } from '../../../stores/files';

import { useOperateinStore } from './../../../stores/operation';
import ReName from './ReName.vue';
import SelectLocal from './SelectLocal.vue';
import DeleteRepo from './DeleteRepo.vue';
import DeleteDialog from './../../../components/files/prompts/DeleteDialog.vue';
import SyncInfo from './SyncInfo.vue';
import { BtDialog } from '@bytetrade/ui';
import { getParams } from '../../../utils/utils';
import { DriveType } from '../../../utils/interface/files';
import InfoDialog from '../prompts/InfoDialog.vue';

import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { notifyFailed } from 'src/utils/notifyRedefinedUtil';

const props = defineProps({
	item: {
		type: Object,
		required: false
	},
	from: {
		type: String,
		require: false,
		default: ''
	},
	isSide: {
		type: Boolean,
		require: false,
		default: false
	},
	origin_id: {
		type: Number,
		required: true
	}
});

const $q = useQuasar();
const dataStore = useDataStore();
const filesStore = useFilesStore();
const route = useRoute();
const router = useRouter();

const emits = defineEmits(['showPopupProxy', 'hideHeaderMenu']);

const { t } = useI18n();

const showPopupProxy = (value: boolean) => {
	if (value) {
		onBeforeShow();
	}

	emits('showPopupProxy', value);
};

const menuStore = useMenuStore();
const operateinStore = useOperateinStore();

const menuList = ref<any[]>([]);
const isElectron = ref($q.platform.is.electron);

const onBeforeShow = async () => {
	if (props.item && props.item.driveType === DriveType.Sync && props.isSide) {
		if (props.isSide && isElectron.value) {
			await sideElectronMenu();
		} else {
			await noLocalMenu();
		}

		await checkShardUser();
	} else {
		await onBeforeShowDrive();
	}
};

const onBeforeShowDrive = async () => {
	menuList.value = popupMenu.filter((e) => {
		return (
			(e.action === OPERATE_ACTION.RENAME &&
				filesStore.selected[props.origin_id] &&
				filesStore.selected[props.origin_id].length === 1) ||
			(e.action === OPERATE_ACTION.DELETE &&
				filesStore.selected[props.origin_id] &&
				filesStore.selected[props.origin_id].length > 0) ||
			e.action === OPERATE_ACTION.ATTRIBUTES
		);
	});
};

const sideElectronMenu = () => {
	const status = menuStore.syncReposLastStatusMap[props.item?.repo_id]
		? menuStore.syncReposLastStatusMap[props.item?.repo_id].status
		: 0;

	if (status == 0) {
		menuList.value = popupMenu.filter(
			(e) => e.requiredSync == undefined || e.requiredSync == false
		);
	} else {
		menuList.value = popupMenu.filter(
			(e) => e.requiredSync == undefined || e.requiredSync == true
		);
	}
};

const noLocalMenu = () => {
	menuList.value = popupMenu.filter((e) => e.requiredSync == undefined);
};

const checkShardUser = () => {
	let selfMenuList = JSON.parse(JSON.stringify(menuList.value));
	let newMenuList: any[] = [];

	const shard_type = props.item && getParams(props.item.path, 'type');

	for (let i = 0; i < selfMenuList.length; i++) {
		const self = selfMenuList[i];
		if (
			self.action === OPERATE_ACTION.SHARE_WITH &&
			props.item?.type &&
			(props.item.type === 'shared' || shard_type === 'shared')
		) {
			continue;
		}

		if (
			self.action === OPERATE_ACTION.EXIT_SHARING &&
			props.item?.type &&
			props.item.type === 'mine'
		) {
			continue;
		}

		if (self.action === OPERATE_ACTION.RENAME) {
			if (
				props.item?.type &&
				(props.item.type === 'shared' || shard_type === 'shared')
			) {
				continue;
			}

			if (
				filesStore.selected[props.origin_id] &&
				filesStore.selected[props.origin_id].length > 1
			) {
				continue;
			}
		}

		if (
			self.action === OPERATE_ACTION.DELETE &&
			props.item?.type &&
			(props.item.type === 'shared' || shard_type === 'shared')
		) {
			continue;
		}

		newMenuList.push(self);
	}

	menuList.value = newMenuList;

	// if (props.item?.shard_user_hide_flag) {
	// 	menuList.value = menuList.value.filter(
	// 		(cell) => cell.name !== 'Exit Sharing'
	// 	);
	// }
};

const handleEvent = async (action: OPERATE_ACTION, e: any) => {
	const path = '/';
	switch (action) {
		case OPERATE_ACTION.SHARE_WITH:
			console.log('shareRepoInfo', props.item);
			filesStore.shareRepoInfo = props.item as any;
			dataStore.showHover('share-internal-dialog');
			break;
		case OPERATE_ACTION.OPEN_LOCAL_SYNC_FOLDER:
			if ($q.platform.is.electron && props.item) {
				window.electron.api.files.openLocalRepo(props.item.id);
			}
			break;
		case OPERATE_ACTION.SYNCHRONIZE_TO_LOCAL:
			showSelectLocal();
			break;
		case OPERATE_ACTION.UNSYNCHRONIZE:
			if ($q.platform.is.electron && props.item) {
				if (
					menuStore.syncReposLastStatusMap[props.item.id] &&
					(menuStore.syncReposLastStatusMap[props.item.id].status ==
						SYNC_STATE.ING ||
						menuStore.syncReposLastStatusMap[props.item.id].status ==
							SYNC_STATE.WAITING ||
						menuStore.syncReposLastStatusMap[props.item.id].status ==
							SYNC_STATE.INIT)
				) {
					notifyFailed(t('Synchronizing, please try again later.'));
				} else {
					window.electron.api.files.repoRemoveSync(props.item.id);
				}
			}
			break;
		case OPERATE_ACTION.SYNC_IMMEDIATELY:
			if ($q.platform.is.electron && props.item) {
				window.electron.api.files.syncRepoImmediately(props.item.id);
			}
			break;

		case OPERATE_ACTION.RENAME:
			showRename(e);
			break;

		case OPERATE_ACTION.DELETE:
			deleteRepo(e);
			break;

		case OPERATE_ACTION.ATTRIBUTES:
			// dataStore.showHover('info');
			syncRepoInfo();

			break;

		case OPERATE_ACTION.EXIT_SHARING:
			deleteShareRepo();
			break;

		default:
			break;
	}
};

const showSelectLocal = () => {
	const jsonItem = JSON.parse(JSON.stringify(props.item));
	$q.dialog({
		component: SelectLocal,
		componentProps: {
			item: jsonItem
		}
	});
};

const showRename = (e: any) => {
	const jsonItem = JSON.parse(JSON.stringify(props.item));

	if (props.from === 'sync' && props.isSide) {
		$q.dialog({
			component: ReName,
			componentProps: {
				item: jsonItem
			}
		});
	} else {
		operateinStore.handleFileOperate(
			props.origin_id,
			e,
			route,
			OPERATE_ACTION.RENAME,
			DriveType.Drive,
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			async (_action: OPERATE_ACTION, _data: any) => {
				//Do nothing
				dataStore.closeHovers();
			}
		);
	}
};

const deleteRepo = async (e: any) => {
	const jsonItem = JSON.parse(JSON.stringify(props.item));

	console.log('jsonItemjsonItem', jsonItem);

	if (props.from === DriveType.Sync) {
		try {
			if (jsonItem.repo_id) {
				const res = await menuStore.fetchShareInfo(jsonItem.repo_id);
				const shared_user_emails_length = res.shared_user_emails.length || 0;

				$q.dialog({
					component: DeleteRepo,
					componentProps: {
						item: jsonItem,
						shared_length: shared_user_emails_length
					}
				});
			} else {
				$q.dialog({
					component: DeleteDialog,
					componentProps: {
						item: jsonItem
					}
				});
			}
		} catch (error) {
			return false;
		}
	} else {
		operateinStore.handleFileOperate(
			props.origin_id,
			e,
			route,
			OPERATE_ACTION.DELETE,
			DriveType.Drive,
			// eslint-disable-next-line @typescript-eslint/no-unused-vars
			async (_action: OPERATE_ACTION, _data: any) => {
				dataStore.closeHovers();
			}
		);
	}
};

const syncRepoInfo = () => {
	try {
		const jsonItem = JSON.parse(JSON.stringify(props.item));
		$q.dialog({
			component: InfoDialog,
			componentProps: {
				fileRes: jsonItem,
				origin_id: props.origin_id
			}
		}).onOk(async () => {
			console.log('ok');
		});
	} catch (error) {
		return false;
	}
};

const deleteShareRepo = async () => {
	const cur_item = props.item;
	if (!cur_item?.repo_id) {
		return false;
	}

	try {
		BtDialog.show({
			title: t('files_popup_menu.exit_sharing'),
			message: t('exit_sharing_message'),
			okStyle: {
				background: 'yellow-default',
				color: '#1F1F1F'
			},
			cancel: true,
			okText: t('base.confirm'),
			cancelText: t('base.cancel')
		})
			.then(async (res: any) => {
				if (res) {
					let path = '';
					if (cur_item?.share_type === 'personal') {
						path = `seahub/api/v2.1/shared-repos/${cur_item.repo_id}/?share_type=${cur_item.share_type}&user=${cur_item.user_email}`;
					} else {
						path = `seahub/api2/beshared-repos/${cur_item?.repo_id}/?share_type=personal&from=${cur_item?.owner_email}`;
					}
					console.log('deleteShareRepo path', path);

					await seahub().deleteRepo(path);

					const curRouteRepoId = getParams(route.fullPath, 'id');
					if (curRouteRepoId === cur_item?.repo_id && cur_item?.owner_email) {
						if (filesStore.backStack[props.origin_id].length > 0) {
							filesStore.back(props.origin_id);
						} else {
							router.push({
								path: operateinStore.defaultPath
							});
						}
					}
					filesStore.getMenu();
				}
			})
			.catch((err: Error) => {
				console.log('click cancel', err);
			});
	} catch (error) {
		return false;
	}
};
</script>
<style lang="scss" scoped>
.popup-item {
	border-radius: 4px;
}
.menuName {
	white-space: nowrap;
}
</style>
