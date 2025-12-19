<template>
	<div
		class="layoutHeader text-subtitle1 row items-center justify-between background-1"
	>
		<div
			class="row items-center justify-between header-content drag-content-header q-pl-md"
		>
			<div
				class="row items-center justify-start items-no-drag"
				style="width: calc(100% - 240px)"
			>
				<NavigationComponent :origin_id="origin_id" />
				<BreadcrumbsComponent :origin_id="origin_id" />
			</div>

			<div
				class="row items-center justify-end items-no-drag"
				style="width: 240px"
			>
				<div
					v-if="
						hideOption &&
						!displayConnectServer &&
						!isSharePageRoot &&
						!isShareApplication
					"
				>
					<q-btn
						v-if="syncPermission === 'rw'"
						class="btn-size-sm btn-no-text btn-no-border q-mr-xs"
						text-color="ink-2"
						icon="sym_r_drive_folder_upload"
					>
						<q-tooltip>{{ t('prompts.upload') }}</q-tooltip>
						<q-menu class="popup-menu bg-background-2">
							<q-list dense padding>
								<q-item
									class="popup-item text-ink-2"
									clickable
									@click="uploadFiles"
									v-close-popup
								>
									<q-item-section>
										{{ t('files_popup_menu.upload_file') }}
									</q-item-section>
								</q-item>

								<q-item
									class="popup-item text-ink-2"
									clickable
									v-close-popup
									@click="uploadFolder"
								>
									<q-item-section>
										{{ t('files_popup_menu.upload_folder') }}
									</q-item-section>
								</q-item>
							</q-list>
						</q-menu>
					</q-btn>

					<q-btn
						v-if="syncPermission === 'rw'"
						class="btn-size-sm btn-no-text btn-no-border q-mr-xs"
						icon="sym_r_create_new_folder"
						text-color="ink-2"
						@click="newFloder"
					>
						<q-tooltip>{{ t('prompts.newDir') }}</q-tooltip>
					</q-btn>

					<q-btn
						class="btn-size-sm btn-no-text btn-no-border q-mr-xs"
						icon="sym_r_more_horiz"
						text-color="ink-2"
						@click="openPopupMenu"
					>
						<q-tooltip>{{ t('buttons.more') }}</q-tooltip>
						<PopupMenu
							:origin_id="origin_id"
							:item="hoverItemAvtive"
							:isSide="false"
							:from="currentDriveType"
							@showPopupProxy="showPopupProxy"
						/>
					</q-btn>
				</div>

				<div
					class="connect-service"
					v-if="displayConnectServer"
					@click="openConnectServer"
				>
					<q-icon name="sym_r_lan" size="16px" />
					<span class="text-ink-2 text-body3 q-ml-xs">{{
						t('files.connect_to_server')
					}}</span>
				</div>
				<div v-if="isSharePageRoot">
					<q-btn
						class="btn-size-sm btn-no-text btn-no-border q-mr-xs"
						icon="sym_r_tune"
						text-color="ink-2"
						@click="filterShowing = !filterShowing"
					>
						<q-tooltip>{{ t('files.Filter') }}</q-tooltip>
					</q-btn>

					<q-menu
						noParentEvent
						:autoClose="false"
						v-model="filterShowing"
						:offset="[0, 5]"
						style="border-radius: 12px"
					>
						<FilterView @close="filterShowing = false" />
					</q-menu>
				</div>

				<div
					class="separator q-mx-md bg-separator"
					style="width: 1px; height: 20px"
				></div>

				<div class="row items-center q-mr-md swithMode">
					<q-btn
						class="btn-size-sm btn-no-text btn-no-border q-mr-xs"
						:class="{ 'bg-btn-bg-pressed': viewMode == 'list' }"
						icon="sym_r_dock_to_right"
						text-color="ink-2"
						@click="switchView('list')"
					/>
					<q-btn
						class="btn-size-sm btn-no-text btn-no-border"
						:class="{ 'bg-btn-bg-pressed': viewMode == 'mosaic' }"
						icon="sym_r_grid_view"
						text-color="ink-2"
						@click="switchView('mosaic')"
					/>
				</div>
			</div>
		</div>

		<TerminusUserHeaderReminder />
	</div>
</template>

<script lang="ts" setup>
import { nextTick, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { common, commonV2, dataAPIs } from '../../../api';
import { useDataStore } from '../../../stores/data';
import { hideHeaderOpt } from './../../../utils/file';
import { bytetrade } from '@bytetrade/core';

import PopupMenu from '../../../components/files/popup/PopupMenu.vue';
import TerminusUserHeaderReminder from './../../../components/common/TerminusUserHeaderReminder.vue';
import BreadcrumbsComponent from './../../../components/files/BreadcrumbsComponent.vue';
import NavigationComponent from './../../../components/files/NavigationComponent.vue';
import { useI18n } from 'vue-i18n';
import { getParams } from '../../../utils/utils';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { SharePermission } from 'src/utils/interface/share';

import { useQuasar } from 'quasar';
import { DriveType } from '../../../utils/interface/files';

import ConnectServerStep1 from './../../../components/files/smb/ConnectServerStep1.vue';
import ConnectServerPath from './../../../components/files/smb/ConnectServerPath.vue';
import ConnectServerStep3 from './../../../components/files/smb/ConnectServerStep3.vue';
import FilterView from './../../../components/files/share/FilterView.vue';

const $q = useQuasar();
const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const Route = useRoute();
const store = useDataStore();
const filesStore = useFilesStore();

const hoverItemAvtive = ref();

const viewMode = ref((store.user && store.user.viewMode) || 'list');

const currentDriveType = ref();
const hideOption = ref(false);
const displayConnectServer = ref(false);
const isSharePageRoot = ref(false);
const syncPermission = ref('rw');
const filterShowing = ref(false);

const isShareApplication = process.env.APPLICATION == 'SHARE';

const { t } = useI18n();

watch(
	() => store.user?.viewMode,
	(newVaule, oldVaule) => {
		if (oldVaule == newVaule) {
			return;
		}
		if (!newVaule) {
			return;
		}
		viewMode.value = newVaule;
	}
);

watch(
	() => [Route.fullPath, filesStore.currentFileList[props.origin_id]],
	async (newVal) => {
		currentDriveType.value = common().formatUrltoDriveType(newVal[0]);
		hideOption.value = hideHeaderOpt(newVal[0]);

		if (currentDriveType.value == DriveType.Share) {
			if (newVal[1] && newVal[1].permission) {
				syncPermission.value =
					newVal[1].permission >= SharePermission.Edit ? 'rw' : '';
			} else {
				syncPermission.value = 'rw';
			}
		} else {
			if (currentDriveType.value === DriveType.Sync) {
				syncPermission.value = getParams(newVal[0], 'p');
			} else {
				syncPermission.value = 'rw';
			}
		}

		if (currentDriveType.value === DriveType.External) {
			displayConnectServer.value = common().displayConnectServer(newVal[0]);
		} else {
			displayConnectServer.value = false;
		}
		if (currentDriveType.value == DriveType.Share) {
			isSharePageRoot.value = commonV2.isShareRootPage(newVal[0]);
		} else {
			isSharePageRoot.value = false;
		}
	},
	{
		immediate: true
	}
);

watch(
	() => filesStore.nodes,
	(newVal) => {
		currentDriveType.value = common().formatUrltoDriveType(Route.fullPath);
		if (currentDriveType.value === DriveType.External) {
			displayConnectServer.value = common().displayConnectServer(
				Route.fullPath
			);
		}
	},
	{}
);

onMounted(async () => {
	nextTick(() => {
		bytetrade.observeUrlChange.childPostMessage({
			type: 'Files'
		});
	});
});

const switchView = async (type: string) => {
	if (type === store.user.viewMode) {
		return false;
	}

	const data = {
		id: store.user.id,
		viewMode: type
	};

	await store.updateUser(data);
};

const uploadFiles = () => {
	filesStore.selectUploadFiles();
};

const uploadFolder = () => {
	filesStore.selectUploadFolder();
};

const newFloder = () => {
	store.showHover('newDir');
};

const showPopupProxy = (value: any) => {
	if (!value) {
		hoverItemAvtive.value = null;
	}
};

const openPopupMenu = async () => {
	if (filesStore.menu[props.origin_id].length < 1) {
		return;
	}

	if (filesStore.selected[props.origin_id].length > 0) {
		hoverItemAvtive.value =
			filesStore.currentFileList[props.origin_id]?.items[
				filesStore.selected[props.origin_id][0]
			];
		return false;
	}

	hoverItemAvtive.value = filesStore.currentFileList[props.origin_id];

	// const driveType = common().formatUrltoDriveType(Route.path);
	// const currentMenus = filesStore.menu[props.origin_id].find(
	// 	(item) => item && item.children && item.children[0]?.driveType === driveType
	// );

	// const activeMenu = currentMenus?.children?.find(
	// 	(item) => item.label === filesStore.activeMenu(props.origin_id).label
	// );

	// const dataAPI = dataAPIs();
	// const res = await dataAPI.getCurrentRepoInfo(Route.fullPath);

	// hoverItemAvtive.value = { ...activeMenu, ...res };
	// console.log('hoverItemAvtive', hoverItemAvtive.value);
};

const openConnectServer = () => {
	$q.dialog({
		component: ConnectServerStep1,
		componentProps: {
			origin_id: props.origin_id
		}
	}).onOk((res) => {
		if (res && res.connectData) {
			return openConnectServerPath(res.connectData, res.paths);
		}

		if (res) {
			$q.dialog({
				component: ConnectServerStep3,
				componentProps: {
					origin_id: props.origin_id,
					smb_url: res.connectUrl,
					hasFavorite: res.hasFavorite
				}
			}).onOk((res) => {
				if (res.connectData) {
					return openConnectServerPath(res.connectData, res.paths);
				}
			});
		}
	});
};

const openConnectServerPath = (connectData, paths) => {
	$q.dialog({
		component: ConnectServerPath,
		componentProps: {
			origin_id: props.origin_id,
			connectData,
			paths
		}
	});
};
</script>

<style lang="scss">
.layoutHeader {
	color: $ink-1;
	padding: 0;
	background-color: $background-1;

	.header-content {
		height: 56px;
		width: 100%;
	}
}

.swithMode {
	.active {
		background-color: $grey-1;
	}
}

.connect-service {
	padding: 2px 8px;
	border-radius: 8px;
	border: 1px solid $input-stroke;
	cursor: pointer;
	&:hover {
		background-color: rgba($color: #000000, $alpha: 0.03);
	}
}
</style>
