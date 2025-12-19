<template>
	<div class="files-page-root">
		<div class="flur1"></div>
		<div class="flur2"></div>
		<Terminus-user-header :title="t('files.files')"> </Terminus-user-header>
		<terminus-scroll-area class="files-content-mobile q-mt-sm">
			<template v-slot:content>
				<bind-terminus-name :border-show="false">
					<template v-slot:success>
						<div class="home-module-title">
							{{ t('files.drive') }}
						</div>
						<div class="module-content q-mt-md">
							<div v-for="(cell, index) in driveMenus" :key="index">
								<terminus-item
									:show-board="false"
									img-bg-classes="bg-background-3"
									:image-path="cell.icon"
									:whole-picture-size="32"
									:icon-size="16"
									@click="seahubAtion(cell.menu, cell.name)"
								>
									<template v-slot:title>
										<div class="text-subtitle1">{{ cell.name }}</div>
									</template>
									<template v-slot:side>
										<q-icon
											name="sym_r_keyboard_arrow_right"
											size="20px"
											color="ink-3"
										/>
									</template>
								</terminus-item>
								<q-separator
									inset
									class="bg-separator"
									v-if="index + 1 < driveMenus.length"
								/>
							</div>
						</div>

						<div class="home-module-title q-mt-lg" v-if="syncMenus.length > 0">
							{{ t('files.sync') }}
						</div>
						<div
							class="module-content q-mt-md bg-background-1"
							v-if="syncMenus.length > 0"
						>
							<div v-for="(cell, index) in syncMenus" :key="index">
								<terminus-item
									:show-board="false"
									img-bg-classes="bg-background-3"
									:icon-name="cell.icon"
									:whole-picture-size="32"
									@click="seahubAtion(cell.menu, cell.name)"
								>
									<template v-slot:title>
										<div class="text-subtitle1">{{ cell.name }}</div>
									</template>
									<template v-slot:side>
										<q-icon
											name="sym_r_keyboard_arrow_right"
											size="20px"
											color="ink-3"
										/>
									</template>
								</terminus-item>
								<q-separator
									inset
									class="bg-separator"
									v-if="index + 1 < syncMenus.length"
								/>
							</div>
						</div>

						<div
							class="home-module-title q-mt-lg"
							v-if="cloudDriveMenus.length > 0"
						>
							{{ t('files.cloud_drive') }}
						</div>
						<div
							class="module-content q-mt-md bg-background-1"
							v-if="cloudDriveMenus.length > 0"
						>
							<div v-for="(cell, index) in cloudDriveMenus" :key="index">
								<terminus-item
									:show-board="false"
									img-bg-classes="bg-background-3"
									:image-path="getAccountIcon(cell)"
									:whole-picture-size="32"
									:icon-size="32"
									contentFrontendClasses="terminus-item-title-part"
									@click="openCloudDrive(cell)"
								>
									<template v-slot:title>
										<div class="text-subtitle1 terminus-text-ellipsis">
											{{ cell.name }}
										</div>
									</template>
									<template v-slot:side>
										<q-icon
											name="sym_r_keyboard_arrow_right"
											size="20px"
											color="ink-3"
										/>
									</template>
								</terminus-item>
								<q-separator
									inset
									class="bg-separator"
									v-if="index + 1 < cloudDriveMenus.length"
								/>
							</div>
						</div>

						<div class="home-module-title q-mt-lg" v-if="shareMenus.length > 0">
							{{ t('files.Shared') }}
						</div>
						<div
							class="module-content q-mt-md bg-background-1"
							v-if="shareMenus.length > 0"
						>
							<div v-for="(cell, index) in shareMenus" :key="index">
								<terminus-item
									:show-board="false"
									img-bg-classes="bg-background-3"
									:icon-name="cell.icon"
									:whole-picture-size="32"
									@click="seahubAtion(cell.menu, cell.name)"
								>
									<template v-slot:title>
										<div class="text-subtitle1">{{ cell.name }}</div>
									</template>
									<template v-slot:side>
										<q-icon
											name="sym_r_keyboard_arrow_right"
											size="20px"
											color="ink-3"
										/>
									</template>
								</terminus-item>
								<q-separator
									inset
									class="bg-separator"
									v-if="index + 1 < syncMenus.length"
								/>
							</div>
						</div>

						<div
							v-if="
								termipassStore &&
								termipassStore.totalStatus &&
								termipassStore.totalStatus.isError == 2
							"
							style="padding-bottom: 60px; width: 100%; height: 1px"
						/>
						<div
							v-else
							style="padding-bottom: 30px; width: 100%; height: 1px"
						/>
					</template>
				</bind-terminus-name>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import TerminusUserHeader from '../../../components/common/TerminusUserHeader.vue';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import { useI18n } from 'vue-i18n';
import { ref, onMounted, computed } from 'vue';
import { MenuItem } from '../../../utils/contact';
import { useRouter } from 'vue-router';
import BindTerminusName from '../../../components/common/BindTerminusName.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { useTermipassStore } from '../../../stores/termipass';
import { FilesIdType, useFilesStore } from '../../../stores/files';
import { useIntegrationStore } from '../../../stores/integration';
import { AccountType, IntegrationAccountMiniData } from '@bytetrade/core';
import integrationService from '../../../services/integration';
import { DriveType } from '../../../utils/interface/files';
import { filesIsV2 } from '../../../api';
import * as filesUtil from '../../../api/files/v2/common/utils';

const { t } = useI18n();
const fileStore = useFilesStore();
const Router = useRouter();
const termipassStore = useTermipassStore();
const integrationStore = useIntegrationStore();

// const driveList = ref();

const driveMenus = ref([
	{
		name: t(`files_menu.${MenuItem.HOME}`),
		icon: 'images/files-home.svg',
		menu: MenuItem.HOME
	},
	{
		name: t(`files_menu.${MenuItem.DOCUMENTS}`),
		icon: 'images/files-document.svg',
		menu: MenuItem.DOCUMENTS
	},
	{
		name: t(`files_menu.${MenuItem.PICTURES}`),
		icon: 'images/files-picture.svg',
		menu: MenuItem.PICTURES
	},
	{
		name: t(`files_menu.${MenuItem.MOVIES}`),
		icon: 'images/files-video.svg',
		menu: MenuItem.MOVIES
	},
	{
		name: t(`files_menu.${MenuItem.DOWNLOADS}`),
		icon: 'images/files-download.svg',
		menu: MenuItem.DOWNLOADS
	},
	{
		name: t(`files_menu.${MenuItem.EXTERNAL}`),
		icon: 'images/file-external.svg',
		menu: MenuItem.EXTERNAL
	}
]);

const seahubAtion = (menu: MenuItem, name?: string) => {
	// const userStore = useUserStore();
	// const termipassStore = useTermipassStore();
	// if (termipassStore.totalStatus?.isError != UserStatusActive.active) {
	// 	notifyFailed(
	// 		t('the_current_status_this_module_cannot_be_accessed', {
	// 			status: termipassStore.totalStatus?.title
	// 		})
	// 	);
	// 	return;
	// }

	// dataStore.updateActiveMenu(menu);

	const query = {
		name: name ? name : menu
	};

	switch (menu) {
		case MenuItem.DATA:
			Router.push({
				path: '/Files/Application/'
			});

			break;

		case MenuItem.CACHE:
			Router.push({
				path: '/Files/AppData/'
			});

			break;

		case MenuItem.MYLIBRARIES:
		case MenuItem.SHAREDWITH:
			fileStore.mobileRepo = {
				path: `/repo/${menu}/`,
				query
			};
			Router.push(fileStore.mobileRepo);
			// openSyncFolder(menu, query);
			break;

		case MenuItem.EXTERNAL:
			fileStore.setBrowserUrl('/Files/External/', DriveType.External);
			break;
		case MenuItem.SHARE:
			// Router.push({
			// 	path: '/Share/'
			// });
			fileStore.setBrowserUrl('/Share/', DriveType.Share);
			break;

		default:
			if (menu === MenuItem.HOME) {
				const url = `/Files/Home/`;
				openDriveFolder(menu, url);
			} else {
				const url = `/Files/Home/${menu}/`;
				openDriveFolder(menu, url);
			}

			break;
	}
};

const openDriveFolder = (menu: string, url: string) => {
	fileStore.setBrowserUrl(url, DriveType.Drive);
};

//0730 hide sync
const syncMenus = computed(() => {
	// if (filesIsV2()) {
	// 	return [];
	// }

	return [
		{
			name: t(`files_menu.${MenuItem.MYLIBRARIES}`),
			icon: 'sym_r_library_books',
			menu: MenuItem.MYLIBRARIES
		}
		// {
		// 	name: t(`files_menu.${MenuItem.SHAREDWITH}`),
		// 	icon: 'sym_r_folder_copy',
		// 	menu: MenuItem.SHAREDWITH
		// }
	];
});

const cloudDriveMenus = computed(() => {
	const supports = integrationStore.clientFilesCloudSupportList();
	return integrationStore.accounts.filter(
		(e) =>
			e.type != AccountType.Space && e.available && supports.includes(e.type)
	);
});

onMounted(async () => {
	integrationStore.getAccount('all');
	fileStore.mobileRepo = undefined;

	fileStore.initIdState(FilesIdType.PAGEID);

	const filesStore = useFilesStore();
	if (filesStore.nodes.length == 0 && filesIsV2()) {
		await filesUtil.fetchNodeList();
	}
});

const getAccountIcon = (data: IntegrationAccountMiniData) => {
	const account = integrationService.getAccountByType(data.type);
	if (!account) {
		return '';
	}
	return `setting/integration/${account.detail.icon}`;
};

const openCloudDrive = async (value: IntegrationAccountMiniData) => {
	const item = {
		...value,
		label: value.name,
		key: value.name,
		icon: '',
		driveType: value.type as any
	};
	const path = await fileStore.formatRepotoPath(item);
	fileStore.setBrowserUrl(path, item.driveType, true);
};

const shareMenus = computed(() => {
	// if (filesIsV2()) {
	// 	return [];
	// }

	return [
		{
			name: t(`files_menu.${MenuItem.SHARE}`),
			icon: 'sym_r_folder_supervised',
			menu: MenuItem.SHARE
		}
	];
});
</script>

<style scoped lang="scss">
.files-page-root {
	height: 100%;
	width: 100%;
	position: relative;
	z-index: 0;

	.flur1 {
		width: 135px;
		height: 135px;
		background: rgba(133, 211, 255, 0.3);
		filter: blur(70px);
		position: absolute;
		right: 10vw;
		top: 10vh;
		z-index: -1;
	}

	.flur2 {
		width: 110px;
		height: 110px;
		background: rgba(217, 255, 109, 0.2);
		filter: blur(70px);
		position: absolute;
		left: 10vw;
		top: 30vh;
		z-index: -1;
	}

	.files-content-mobile {
		height: calc(100% - 56px);
		width: 100%;

		padding-left: 20px;
		padding-right: 20px;

		.module-content {
			border: 1px solid $separator;
			border-radius: 20px;
			width: 100%;
		}
	}

	.func-menu {
		width: 40px;
		height: 40px;
		position: relative;
		.tranfering {
			position: absolute;
			top: 2px;
			right: 2px;
			display: inline-block;
			width: 8px;
			height: 8px;
			border-radius: 4px;
			background-color: $negative;
		}
	}
}
</style>
