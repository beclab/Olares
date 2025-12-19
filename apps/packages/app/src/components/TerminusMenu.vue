<template>
	<terminus-drag-node
		class="contain-bar"
		:class="
			deviceStore.isLandscape ? 'contain-bar-landscape' : 'contain-bar-portrait'
		"
	>
		<div
			class="contain-header cursor-pointer items-center"
			:style="
				$q.platform.is.electron && $q.platform.is.win ? '' : 'margin-top:24px;'
			"
			@dblclick.stop
			:class="
				deviceStore.isLandscape
					? 'contain-header-landscape'
					: 'row items-center justify-center'
			"
		>
			<div class="avator">
				<TerminusAvatar
					v-if="current_user?.id"
					:info="userStore.terminusInfo()"
					:size="40"
				/>
			</div>
			<div class="userinfo q-ml-sm" v-if="deviceStore.isLandscape">
				<div class="text-subtitle1 text-left">
					{{ current_user?.local_name }}
				</div>
				<div class="text-overline text-left row items-start">
					<TerminusUserStatus @super-action="showing = true" />
				</div>
			</div>
			<q-menu
				v-model="showing"
				:offset="[-45, -50]"
				style="border-radius: 12px"
			>
				<TerminusAdmin
					@switchAccount="handleSwitchAccount"
					@handleSettings="handleSettings"
				/>
			</q-menu>
		</div>

		<div
			class="contain-search q-mt-sm"
			ref="searchBox"
			v-if="deviceStore.isLandscape"
		>
			<q-input
				dense
				stack-label
				borderless
				readonly
				class="search_itme"
				v-model="searchVal"
				@click.stop
				@dblclick.stop
				debounce="500"
				@update:model-value="updateSearch"
				:placeholder="t('search')"
				input-style="color: rgba(92, 85, 81, 1); height: 30px !important; line-height: 30px;"
			>
				<template v-slot:prepend>
					<q-icon class="search_icon" name="search" size="16px" />
				</template>
				<template v-slot:append>
					<q-icon
						v-if="searchVal"
						name="sym_r_cancel"
						size="16px"
						@click="clearSearch"
						class="search_clean cursor-pointer"
					/>
					<span class="command">
						<q-icon
							v-if="isMac"
							name="sym_r_keyboard_command_key"
							size="12px"
						/>
						<span v-else class="text-overline">Ctrl</span>
						<span v-if="!searchVal" class="text-overline"> + K</span>
					</span>
				</template>
			</q-input>

			<div
				@click.stop="openSearch"
				style="
					position: absolute;
					top: 0;
					left: 0;
					width: 100%;
					height: 100%;
					cursor: pointer;
				"
			></div>
		</div>
		<div v-else class="contain-search-portrait row items-center justify-center">
			<!-- <q-icon name="vector" /> -->
			<div class="contain-search-bg row items-center justify-center">
				<q-icon name="sym_r_search" size="18px" color="text-ink-2" />
			</div>
		</div>

		<q-list
			class="menuList"
			:style="
				deviceStore.isLandscape ? 'margin-top: 20px;' : 'margin-top: 10px'
			"
		>
			<template v-for="menu in filterLayoutMenu" :key="menu.name">
				<q-item
					clickable
					v-ripple
					v-close-popup
					dense
					@dblclick.stop
					@click.stop="handleActive(menu)"
					:active="menu.identify === menuStore.terminusActiveMenu"
					class="q-px-md q-mb-xs q-py-none text-grey-8 text-body1"
					active-class="text-ink-1 bg-background-2"
					style="border-radius: 8px"
					:class="
						deviceStore.isLandscape
							? 'row items-center justify-start '
							: 'column items-center justify-center '
					"
					:style="deviceStore.isLandscape ? 'height: 40px' : ' height: 50px;'"
				>
					<q-img
						class="q-mr-sm"
						style="width: 20px; height: 20px"
						:src="
							require(`../assets/layout/${
								menu.identify === menuStore.terminusActiveMenu
									? $q.dark.isActive
										? menu.icon_active_dark
										: menu.icon_active
									: menu.icon
							}.svg`)
						"
					/>
					<div
						class="trans-title row items-center justify-between"
						:class="
							menu.identify === menuStore.terminusActiveMenu
								? 'title-active text-body1'
								: 'title-normal text-body1'
						"
					>
						{{ t(deviceStore.isLandscape ? menu.name : menu.short_name) }}
						<span
							class="trans-num text-subtitle3 text-white"
							v-if="menu.name === 'transmission.title' && transNum"
							>{{ transNum }}</span
						>
					</div>
				</q-item>
			</template>
		</q-list>
	</terminus-drag-node>
</template>
<script lang="ts" setup>
import { ref, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useRouter } from 'vue-router';
import { useMenuStore } from '../stores/menu';
import { useUserStore } from '../stores/user';
import { LayoutMenu, LayoutMenuIdetify } from '../utils/constants';

import TerminusAdmin from './TerminusAdmin.vue';
import SwitchAccount from './SwitchAccount.vue';
import TerminusDragNode from './common/TerminusDragNode.vue';
import TerminusUserStatus from './common/TerminusUserStatus.vue';
import { getAppPlatform } from '../application/platform';
// import UserManagentDialog from '../pages/Pad/account/UserManagentDialog.vue';
import SettingsDialog from '../pages/Pad/settings/SettingsDialog.vue';
import { useDeviceStore } from '../stores/device';
import { useI18n } from 'vue-i18n';

import { useTransfer2Store } from '../stores/transfer2';

const emits = defineEmits(['openSearch']);

const $q = useQuasar();
const menuStore = useMenuStore();
const router = useRouter();
const userStore = useUserStore();
const filterLayoutMenu = ref(LayoutMenu);

const transferStore = useTransfer2Store();

const { t } = useI18n();

const current_user = ref(userStore.current_user);
const searchVal = ref();
const isMac = $q.platform.is.mac;

const transNum = computed(() => {
	const downloadingNum = transferStore.downloading.length;
	const uploadingNum = transferStore.uploading.length;
	const cloudingNum = transferStore.clouding.length;
	const copyingNum = transferStore.copying.length;

	if (downloadingNum + uploadingNum + cloudingNum + copyingNum > 99) {
		return '99+';
	}
	return downloadingNum + uploadingNum + cloudingNum + copyingNum;
});

const updateSearch = (value: any) => {
	searchVal.value = value;
	filterLayoutMenu.value = fuzzySearch(value);
};

const fuzzySearch = (keyword) => {
	return LayoutMenu.filter((item) =>
		item.name.toLowerCase().includes(keyword.toLowerCase())
	);
};

const clearSearch = () => {
	searchVal.value = '';
	filterLayoutMenu.value = LayoutMenu;
};

const handleActive = (menu: {
	name: string;
	icon: string;
	icon_active: string;
	identify: string;
	path: string;
}) => {
	menuStore.pushTerminusMenuCache(menu.identify);

	if (menu.identify === LayoutMenuIdetify.VAULT) {
		menuStore.currentItem = 'All Vaults';
	}
	router.push({
		path: menu.path
	});
};

const handleSwitchAccount = () => {
	$q.dialog({
		component: SwitchAccount
	});
};

// const handleAccountCenter = () => {
// 	if (isPad.value) {
// 		showing.value = false;
// 		$q.dialog({
// 			component: UserManagentDialog
// 		});
// 		return;
// 	}
// 	menuStore.pushTerminusMenuCache(LayoutMenuIdetify.ACCOUNT_CENTER);
// 	router.push({
// 		path: '/accountCenter'
// 	});
// };

const handleSettings = () => {
	if (isPad.value) {
		showing.value = false;
		$q.dialog({
			component: SettingsDialog
		});
		return;
	}
	menuStore.pushTerminusMenuCache(LayoutMenuIdetify.SYSTEM_SETTINGS);
	router.push({
		path: '/systemSettings'
	});
};

const openSearch = () => {
	emits('openSearch');
};

const showing = ref(false);

const isPad = ref(getAppPlatform() && getAppPlatform().isPad);
const deviceStore = useDeviceStore();
</script>

<style lang="scss" scoped>
.contain-bar {
	height: 100vh;
	text-align: center;
	position: relative;
	padding-top: 12px;
	padding-left: 12px;
	padding-right: 12px;

	.contain-header {
		width: 100%;
		height: 62px;

		.avator {
			width: 40px;
			height: 40px;
			border-radius: 20px;
			overflow: hidden;
		}

		.userinfo {
			width: 116px;

			div {
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
			}
		}
	}

	.contain-header-landscape {
		display: flex;
		align-items: center;
		justify-content: flex-start;
		-webkit-app-region: no-drag;
	}

	.contain-search {
		position: relative;
		-webkit-app-region: no-drag;

		.search_itme {
			width: 100%;
			height: 32px !important;
			line-height: 32px !important;
			border-radius: 8px;
			font-size: map-get($map: $body2, $key: size) !important;
			padding-left: 8px;
			padding-right: 8px;
			border: 1px solid $separator;
			box-sizing: border-box;
			// background: $grey-1;
			position: relative;

			.search_icon {
				margin-bottom: 8px;
			}

			.search_clean {
				margin-bottom: 8px;
			}

			.command {
				display: inline-block;
				height: 32px;
				position: absolute;
				top: 0;
				right: 0;
				display: flex;
				align-items: center;
				justify-center: center;
				color: $grey-5;

				.k {
					height: 16px;
					line-height: 16px;
					display: inline-block;
					color: $grey-5;
				}
			}
		}
	}

	.contain-search-portrait {
		width: 100%;
		height: 40px;
		// background: $background-alpha;
		.contain-search-bg {
			width: 40px;
			height: 100%;
			background: $background-alpha;
			border-radius: 50%;
		}
	}

	.menuList {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		// background-color: red;

		margin-top: 20px;
		.title-active {
			color: $ink-1;
		}

		.title-normal {
			color: $ink-2;
		}

		.trans-title {
			width: calc(100% - 20px);
		}

		.trans-num {
			display: inline-block;
			background-color: $negative;
			height: 16px;
			border-radius: 8px;
			padding: 0 8px;
		}
	}
}

.contain-bar-landscape {
	width: 180px;
}
.contain-bar-portrait {
	width: 70px;
}
</style>
