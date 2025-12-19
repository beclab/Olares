<template>
	<wise-restore-view />
	<q-layout view="hHh Lpr fFf" class="layout layout-app">
		<div class="main-layout">
			<q-drawer
				v-model="configStore.leftDrawerOpen"
				@update:model-value="updateLeftDrawer"
				show-if-above
				bordered
				height="100%"
				:width="240"
			>
				<bt-scroll-area style="height: calc(100vh - 52px)">
					<bt-menu
						active-class="my-active-link"
						:items="items"
						:show-theme-toggle="false"
						v-model="configStore.menuChoice.type"
						@select="onMenuItemChange"
					>
						<template #extra-trend>
							<div id="hotkey" style="display: none">
								<bt-hot-key-icon :hotkey="WISE_HOTKEY.MENU.TREND" />
							</div>
						</template>

						<template
							v-for="(item, index) in filterStore.filterList.filter(
								(data) => data.pin
							)"
							:key="item.id"
							v-slot:[`extra-${item.id}`]
						>
							<div id="hotkey" style="display: none">
								<bt-hot-key-icon
									:hotkey="WISE_HOTKEY.MENU[`VIEWS` + (index + 1)]"
								/>
							</div>
						</template>

						<template #extra-history>
							<div id="hotkey" style="display: none">
								<bt-hot-key-icon :hotkey="WISE_HOTKEY.MENU.RECENTLY" />
							</div>
						</template>
					</bt-menu>
				</bt-scroll-area>

				<q-btn
					class="q-ml-sm btn-size-md btn-no-text btn-no-border"
					icon="sym_r_add_circle"
					color="ink-2"
					outline
					no-caps
				>
					<bt-tooltip :label="t('dialog.add')" />
					<bt-popup style="width: 200px">
						<bt-popup-item
							v-close-popup
							icon="sym_r_upload"
							:title="t('dialog.upload')"
							:hotkey="WISE_HOTKEY.ADD.UPLOAD"
							@on-item-click="addUpload"
						/>

						<bt-popup-item
							v-if="configStore.uploadCookiesOpen"
							v-close-popup
							icon="sym_r_public"
							:title="t('dialog.upload_cookie')"
							:hotkey="WISE_HOTKEY.ADD.COOKIES"
							@on-item-click="addCookies"
						/>

						<bt-popup-item
							v-if="configStore.uploadLinksOpen"
							v-close-popup
							icon="sym_r_public"
							:title="t('dialog.batch_add_link')"
							:hotkey="WISE_HOTKEY.ADD.LINKS"
							@on-item-click="addEntries"
						/>

						<bt-popup-item
							v-close-popup
							icon="sym_r_public"
							:title="t('dialog.add_link')"
							:hotkey="WISE_HOTKEY.ADD.URL"
							@on-item-click="addEntry"
						/>
					</bt-popup>
				</q-btn>

				<q-btn
					class="q-ml-sm btn-size-md btn-no-text btn-no-border"
					icon="sym_r_settings"
					color="ink-2"
					outline
					no-caps
				>
					<bt-tooltip :label="t('setting')" />
					<bt-popup style="width: 200px">
						<bt-popup-item
							v-close-popup
							icon="sym_r_swap_vert"
							:title="t('main.transmission')"
							:hotkey="WISE_HOTKEY.SETTING.TRANSACTION"
							@on-item-click="configStore.setMenuType(MenuType.Transmission)"
						/>
						<bt-popup-item
							v-close-popup
							icon="sym_r_sell"
							:title="t('base.tags')"
							:hotkey="WISE_HOTKEY.SETTING.TAG"
							@on-item-click="configStore.setMenuType(MenuType.Tags)"
						/>
						<bt-popup-item
							v-close-popup
							icon="sym_r_rss_feed"
							:title="t('main.rss_feeds')"
							:hotkey="WISE_HOTKEY.SETTING.RSS"
							@on-item-click="configStore.setMenuType(MenuType.RSS_Feeds)"
						/>
						<bt-popup-item
							v-close-popup
							icon="sym_r_grid_view"
							:title="t('main.filtered_views')"
							:hotkey="WISE_HOTKEY.SETTING.VIEW"
							@on-item-click="configStore.setMenuType(MenuType.Filtered_Views)"
						/>
						<bt-popup-item
							v-if="configStore.recommendationOpen"
							v-close-popup
							icon="sym_r_featured_play_list"
							:title="t('main.recommendations')"
							:hotkey="WISE_HOTKEY.SETTING.RECOMMEND"
							@on-item-click="configStore.setMenuType(MenuType.Recommend)"
						/>
						<bt-popup-item
							v-close-popup
							icon="sym_r_settings_applications"
							:title="t('main.preferences')"
							:hotkey="WISE_HOTKEY.SETTING.PREFERENCE"
							@on-item-click="configStore.setMenuType(MenuType.Preferences)"
						/>
					</bt-popup>
				</q-btn>
			</q-drawer>

			<q-page-container>
				<div class="container">
					<router-view v-slot="{ Component }">
						<transition
							:name="transitionName"
							class="bg-background-1"
							:style="`position: absolute;overflow:scroll;-ms-overflow-style:none;scrollbar-width:none;width:calc(100% - ${
								configStore.leftDrawerOpen ? '240px' : '0px'
							} - ${configStore.rightDrawerOpen ? '320px' : '0px'})`"
						>
							<keep-alive max="10" :exclude="keepAliveExclude">
								<component
									:is="Component"
									style="overflow-y: hidden"
									:key="route.fullPath"
								/>
							</keep-alive>
						</transition>
					</router-view>

					<files-uploader
						accept="video/*, audio/*, application/pdf, application/epub+zip, application/vnd.amazon.ebook"
						:auto-bind-resumable="false"
						@files-update="filesUpdate"
					/>
				</div>
			</q-page-container>

			<q-drawer
				side="right"
				:model-value="configStore.rightDrawerOpen"
				bordered
				style="overflow: hidden"
				:width="configStore.rightDrawerOpen ? 320 : 0"
				@update:model-value="(value) => (configStore.rightDrawerOpen = value)"
			>
				<right-drawer-layout />
			</q-drawer>
		</div>
	</q-layout>
</template>

<script lang="ts" setup>
import TransferUploadAddDialog from '../pages/Electron/Transfer/TransferUploadAddDialog.vue';
import MultiTextDialog from 'src/components/rss/dialog/MultiTextDialog.vue';
import FilesUploader from '../pages/Files/common-files/FilesUploader.vue';
import MultiAddDialog from 'src/components/rss/dialog/MultiAddDialog.vue';
import WiseRestoreView from 'src/components/rss/WiseRestoreView.vue';
import RightDrawerLayout from '../pages/Wise/RightDrawerLayout.vue';
import BtHotKeyIcon from 'src/components/base/BtHotKeyIcon.vue';
import BtPopupItem from '../components/base/BtPopupItem.vue';
import BtTooltip from '../components/base/BtTooltip.vue';
import BtPopup from '../components/base/BtPopup.vue';
import Waiter from '../utils/simpleWaiter';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { onBeforeRouteUpdate, useRoute, useRouter } from 'vue-router';
import { MenuType, SupportDetails, TabType } from '../utils/rss-menu';
import { FilePath, useFilesStore } from '../stores/files';
import HotkeyManager from 'src/directives/hotkeyManager';
import { WISE_HOTKEY } from 'src/directives/wiseHotkey';
import { useFilterStore } from '../stores/rss-filter';
import { useConfigStore } from '../stores/rss-config';
import { DriveType } from 'src/utils/interface/files';
import { useRssStore } from 'src/stores/rss';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { dataAPIs } from '../api';

const keepAliveExclude = ref('EntryReadingPage');
const configStore = useConfigStore();
const fileStore = useFilesStore();
const $q = useQuasar();
const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const transitionName = ref();
const filesRef = ref();
const targetRef = ref();
const filterStore = useFilterStore();
const rssStore = useRssStore();
const rssWait = new Waiter<string>();

const items = computed(() => {
	return configStore.recommendationOpen
		? [
				{
					label: t('main.subscription'),
					key: 'Subscription',
					children: [
						{
							label: t('main.for_you'),
							key: MenuType.Trend,
							icon: 'sym_r_volunteer_activism'
						}
					]
				},
				{
					label: t('main.library'),
					key: 'Wise',
					children: getCustomMenu()
				},
				{
					label: t('main.manage'),
					key: 'Managed',
					children: [
						{
							label: t('main.recently_read'),
							key: MenuType.History,
							icon: 'sym_r_history'
						}
					]
				}
		  ]
		: [
				{
					label: t('main.library'),
					key: 'Wise',
					children: getCustomMenu()
				},
				{
					label: t('main.manage'),
					key: 'Managed',
					children: [
						{
							label: t('main.recently_read'),
							key: MenuType.History,
							icon: 'sym_r_history'
						}
					]
				}
		  ];
});

watch(
	() => configStore.menuInited,
	() => {
		if (configStore.menuInited) {
			HotkeyManager.registerHotkeys({
				[WISE_HOTKEY.MENU.RECENTLY]: () =>
					configStore.setMenuType(MenuType.History),
				[WISE_HOTKEY.ADD.UPLOAD]: () => addUpload(),
				[WISE_HOTKEY.ADD.COOKIES]: () => addCookies(),
				[WISE_HOTKEY.ADD.LINKS]: () => addEntries(),
				[WISE_HOTKEY.ADD.URL]: () => addEntry(),
				[WISE_HOTKEY.SETTING.TRANSACTION]: () =>
					configStore.setMenuType(MenuType.Transmission),
				[WISE_HOTKEY.SETTING.TAG]: () => configStore.setMenuType(MenuType.Tags),
				[WISE_HOTKEY.SETTING.RSS]: () =>
					configStore.setMenuType(MenuType.RSS_Feeds),
				[WISE_HOTKEY.SETTING.VIEW]: () =>
					configStore.setMenuType(MenuType.Filtered_Views),
				// [WISE_HOTKEY.SETTING.RECOMMEND]: () =>
				// 	configStore.setMenuType(MenuType.Recommend),
				[WISE_HOTKEY.SETTING.PREFERENCE]: () =>
					configStore.setMenuType(MenuType.Preferences)
			});

			HotkeyManager.registerHotkeys(
				{
					[WISE_HOTKEY.TAB.NEXT]: () => {
						configStore.nextTab();
					},
					[WISE_HOTKEY.TAB.PRE]: () => {
						configStore.preTab();
					}
				},
				[MenuType.Trend, MenuType.Custom]
			);
		}
	},
	{
		immediate: true
	}
);

function getCustomMenu(): {
	label: string;
	key: string;
	icon: string;
	params: any;
	system: boolean;
}[] {
	const sortedList = filterStore.filterList
		.filter((item) => item.pin)
		.sort((a, b) => {
			return a.serial_no - b.serial_no;
		});
	if (sortedList.length > 0) {
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS1);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS2);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS3);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS4);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS5);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS6);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS7);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS8);
		HotkeyManager.unbind(WISE_HOTKEY.MENU.VIEWS9);
		sortedList.forEach((item, index) => {
			if (index > -1 && index < 9) {
				HotkeyManager.bind({
					key: (index + 1).toString(),
					handler: () => {
						configStore.setMenuType(item.id, { filterId: item.id });
					}
				});
			}
		});

		return sortedList.map((item) => {
			const menu: any = {
				label: item.name,
				key: `${item.id}`,
				icon: item.icon ? item.icon : 'sym_r_filter',
				params: {
					filterId: item.id
				},
				system: item.system,
				count: undefined
			};

			if (item.system) {
				menu.label = t(`main.${menu.label}`);
			}

			if (item && item.showbadge) {
				const size = filterStore.unseenMap[item.id];
				menu.count = size === 0 ? undefined : size;
			} else {
				menu.count = undefined;
			}
			return menu;
		});
	}
	return [];
}

const updateLeftDrawer = (show: boolean) => {
	configStore.leftDrawerOpen = show;
};

onBeforeRouteUpdate((to, from, next) => {
	transitionName.value = '';
	next();
});

watch(
	() => route.path,
	(to, from) => {
		if (router.options.history.state) {
			if (route.params.action && route.params.action === 'back') {
				transitionName.value = 'slide-right';
				return;
			} else if (route.params.id) {
				transitionName.value = 'slide-left';
				return;
			}
		}

		const toPathSegments = to.split('/').filter(Boolean);
		const fromPathSegments = from.split('/').filter(Boolean);

		if (fromPathSegments.length < toPathSegments.length) {
			transitionName.value = 'slide-left';
		} else if (fromPathSegments.length > toPathSegments.length) {
			transitionName.value = 'slide-right';
		} else {
			transitionName.value = '';
		}
		// console.log(toPathSegments);
		// console.log(fromPathSegments);
		// console.log(transitionName.value);
	}
);

watch(
	() => configStore.menuChoice,
	() => {
		if (configStore.menuChoice.params.break) {
			return;
		}
		if (
			configStore.menuChoice.params &&
			configStore.menuChoice.params.filterId
		) {
			router.push({
				name: MenuType.Custom,
				params: configStore.menuChoice.params
			});
		} else {
			router.push({
				name: configStore.menuChoice.type,
				params: configStore.menuChoice.params
			});
		}
	},
	{
		deep: true
	}
);

onMounted(async () => {
	console.log('main onMounted');
	console.log(route.path);
	configStore.setThemeSetting(configStore.themeSetting);
	restoreMenu();
});

const restoreMenu = () => {
	if (route.name === 'WiseMain') {
		rssWait.waitForCondition(
			() => filterStore.inited,
			() => {
				const pinList = filterStore.filterList.filter((item) => item.pin);
				if (pinList.length > 0) {
					configStore.setMenuType(pinList[0].id, { filterId: pinList[0].id });
				} else {
					configStore.setMenuType(MenuType.History);
				}

				configStore.menuInited = true;
			},
			500
		);
	} else if (route.name === MenuType.Trend) {
		configStore.setMenuType(MenuType.Trend, rssStore.support_algorithm[0].id);
	} else if (route.name === MenuType.Entry) {
		console.log(route.path);
		const array = route.path.split('/');
		if (array.length > 0) {
			const type = array[1];
			if (SupportDetails.includes(type)) {
				console.log(type);
				configStore.setMenuType(type, {
					break: true,
					id: array[2]
				});
				configStore.menuInited = true;
			} else {
				rssWait.waitForCondition(
					() => filterStore.inited,
					() => {
						console.log('do wait');
						configStore.setMenuType(route.params.path, {
							break: true,
							id: array[2]
						});
						configStore.menuInited = true;
					},
					500
				);
			}
		}
	} else if (route.name === MenuType.Custom) {
		rssWait.waitForCondition(
			() => filterStore.inited,
			() => {
				console.log('do wait');
				configStore.setMenuType(route.params.filterId, route.params);
				configStore.menuInited = true;
			},
			500
		);
	} else {
		configStore.setMenuType(route.name, route.params);
		configStore.menuInited = true;
	}
};

onBeforeUnmount(() => {
	if (rssWait) {
		rssWait.clear();
	}
});

function onMenuItemChange(data: { key; item }) {
	configStore.setMenuType(data.key, data.item?.params);
}

const addEntry = () => {
	$q.dialog({
		component: MultiAddDialog
	});
};

const addEntries = () => {
	$q.dialog({
		component: MultiTextDialog,
		componentProps: {
			title: 'dialog.batch_add_link',
			label: 'dialog.Add links'
		}
	});
};

const addCookies = () => {
	$q.dialog({
		component: MultiTextDialog,
		componentProps: {
			title: 'dialog.upload_cookie',
			label: 'dialog.Add cookies',
			link: false
		}
	});
};

const addUpload = () => {
	const dataAPI = dataAPIs();
	dataAPI.uploadFiles();
};

const filesUpdate = (event: any) => {
	filesRef.value = event.target.files;
	targetRef.value = event;
	showUploadDialog();
};

const showUploadDialog = () => {
	if (!filesRef.value) {
		return;
	}
	$q.dialog({
		component: TransferUploadAddDialog,
		componentProps: {
			files: filesRef.value,
			origins: [DriveType.Drive]
		}
	}).onOk((fileSavePath: FilePath) => {
		fileStore.uploadSelectFile(targetRef.value, fileSavePath);
		configStore.setMenuType(MenuType.Transmission);
		configStore.setMenuTab(TabType.Upload);
	});
};
</script>

<style lang="scss" scoped>
.layout-app {
	perspective: 500;
	-webkit-perspective: 500;
}

.main-layout {
	width: 100vw;
	height: 100vh;

	.bottom-menu-root {
		height: 40px;
		margin-left: 12px;
	}

	.bottom-menu {
		font-size: 12px;
		line-height: 12px;
		font-family: 'Roboto';
		font-style: normal;
		padding: 12px;
	}

	.container {
		height: 100%;
		width: 100%;
	}
}

.rotate {
	animation: aniRotate 0.8s linear infinite;

	&:hover {
		background: transparent !important;
	}
}

@keyframes aniRotate {
	0% {
		transform: rotate(0deg);
	}

	50% {
		transform: rotate(180deg);
	}

	100% {
		transform: rotate(360deg);
	}
}

.main-layout ::v-deep(.my-active-link) {
	color: $orange-default;
	background-color: $orange-soft;
}

.main-layout ::v-deep(.bt-badgecount-container) {
	color: $orange-default;
	background-color: $background-hover;
}

.main-layout ::v-deep(.my-active-link .bt-badgecount-container) {
	color: $orange-default;
	background-color: $background-2;
}

.slide-left-enter-active,
.slide-left-leave-active,
.slide-right-enter-active,
.slide-right-leave-active {
	will-change: transform;
	transition: transform 0.3s;
	height: 100%;
	width: 100%;
	top: 0;
	position: absolute;
	backface-visibility: hidden;
	perspective: 1000;
}

.slide-left-leave-to {
	transform: translateX(-100%);
}

.slide-left-enter-from {
	transform: translateX(100%);
}

.slide-right-leave-to {
	transform: translateX(100%);
}

.slide-right-enter-from {
	transform: translateX(-100%);
}
</style>

<style lang="scss">
.menu-item-wrapper:hover #hotkey {
	display: block !important;
}
</style>
